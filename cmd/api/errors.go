package main

import (
	"fmt"
	"net/http"

	"log/slog"
)

func (app *application) internalError(w http.ResponseWriter, r *http.Request, error error) error {
	slog.Error("internal server error, path: %s; method: %s, error:", fmt.Sprintf("%s", r.URL), r.Method, error)
	return app.writeJsonError(w, http.StatusInternalServerError, "something went wrong")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, error error, customErrorMsg string) error {
	slog.Error("bad request error, path: %s; method: %s, error:", fmt.Sprintf("%s", r.URL), r.Method, error)
	if customErrorMsg == "" {
		customErrorMsg = error.Error()
	}
	return app.writeJsonError(w, http.StatusBadRequest, customErrorMsg)
}
