package main

import (
	"errors"
	"net/http"

	"github.com/9thDuck/chat_go.git/internal/store"
)

const (
	userCtxKey        ctxKey = "user"
	userIDCtxKey      ctxKey = "userID"
	otherUserIDCtxKey ctxKey = "otherUserID"
)

func getUserFromCtx(r *http.Request) *store.User {
	data := r.Context().Value(userCtxKey)
	user, ok := data.(*store.User)
	if !ok {
		return nil
	}
	return user
}

func deleteCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
func (app *application) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from context that was set by middleware
	otherUserID := getOtherUserIDFromCtx(r)

	// Log for debugging
	app.logger.Infow("User ID from context", "otherUserID", otherUserID)

	if otherUserID == 0 {
		app.badRequestError(w, r, errors.New("user ID is required"), "")
		return
	}

	ctx := r.Context()
	user, err := app.getUser(ctx, otherUserID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err, "")
		default:
			app.internalError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalError(w, r, err)
		return
	}
}
func getOtherUserIDFromCtx(r *http.Request) int64 {
	val := r.Context().Value(otherUserIDCtxKey)
	if val == nil {
		return 0
	}

	id, ok := val.(int64)
	if !ok {
		return 0
	}

	return id
}
