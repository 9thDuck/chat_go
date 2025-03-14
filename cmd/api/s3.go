package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) getPresignedS3URLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	objectKey := chi.URLParam(r, "objectKey")
	if objectKey == "" {
		app.badRequestError(w, r, nil, "key can't be empty")
		return
	}

	operation := r.URL.Query().Get("operation")
	var presignedURL string
	var err error

	switch operation {
	case "upload":
		presignedURL, err = app.cloud.PreSigner.Create(ctx, app.config.cloud.s3.bucketName, objectKey, 60)
	case "download", "":
		presignedURL, err = app.cloud.PreSigner.Get(ctx, app.config.cloud.s3.bucketName, objectKey, 60)
	default:
		app.badRequestError(w, r, nil, "invalid operation type")
		return
	}

	if err != nil {
		app.internalError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, presignedURL); err != nil {
		app.internalError(w, r, err)
		return
	}
}
