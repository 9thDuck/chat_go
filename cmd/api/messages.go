package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/9thDuck/chat_go.git/cmd/api/ws"
	"github.com/9thDuck/chat_go.git/internal/store"
)

type createMessagePayload struct {
	Content     string   `json:"content" validate:"required,min=1,max=1000"`
	Attachments []string `json:"attachments" validate:"omitempty,max=10,dive,max=255"`
}

const receiverIDCtxKey ctxKey = "receiverID"
const messageCreationPayloadCtxKey ctxKey = "messageCreationPayload"

func (app *application) getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	pagination := getPaginationOptionsFromCtx(r)

	messages, total, err := app.store.Messages.Get(r.Context(), user.ID, pagination)
	if err != nil {
		app.internalError(w, r, err)
		return
	}

	if messages != nil {
		app.generateSignedURLsForAttachments(r.Context(), messages)
	}

	app.jsonResponse(w, http.StatusOK, &paginatedEnvelope{
		Records:      messages,
		TotalRecords: total,
	})
}

func (app *application) generateSignedURLsForAttachments(ctx context.Context, messages *[]store.Message) {
	if messages == nil {
		return
	}

	for i := range *messages {
		message := &(*messages)[i]
		if message.Attachments == nil || len(*message.Attachments) == 0 {
			continue
		}

		signedAttachments := make([]string, len(*message.Attachments))
		for j, path := range *message.Attachments {
			signedURL, err := app.cloud.PreSigner.Get(ctx, app.config.cloud.s3.bucketName, path, 3600)
			if err != nil {
				app.logger.Errorw("Failed to generate signed URL", "path", path, "error", err)
			} else {
				signedAttachments[j] = signedURL
			}
		}
		message.Attachments = &signedAttachments
	}
}

func (app *application) createMessageHandler(w http.ResponseWriter, r *http.Request) {
	receiverID := getReceiverIDFromCtx(r)
	payload := getMessagePayloadFromCtx(r)

	user := getUserFromCtx(r)

	message := store.Message{
		SenderID:    user.ID,
		ReceiverID:  receiverID,
		Content:     payload.Content,
		Attachments: &[]string{},
		IsDelivered: false,
		IsRead:      false,
		Version:     1,
		Edited:      false,
	}

	if payload.Attachments != nil {
		message.Attachments = &payload.Attachments
	}

	err := app.store.Messages.Create(r.Context(), &message)

	switch err {
	case nil:
		app.jsonResponse(w, http.StatusCreated, &message)
		jsonMessage, err := json.Marshal(ws.MessageEvent{
			Message: message,
			Type:    ws.EVENT_MESSAGE,
		})
		if err != nil {
			app.logger.Errorw("Failed to marshal message", "error", err)
		}
		if done := app.socketHub.WriteToClient(receiverID, jsonMessage); done {
			app.store.Messages.Delete(r.Context(), message.ID)
		}
		return
	default:
		app.internalError(w, r, err)
		return
	}
}

func getReceiverIDFromCtx(r *http.Request) int64 {
	return r.Context().Value(receiverIDCtxKey).(int64)
}

func getMessagePayloadFromCtx(r *http.Request) *createMessagePayload {
	return r.Context().Value(messageCreationPayloadCtxKey).(*createMessagePayload)
}
