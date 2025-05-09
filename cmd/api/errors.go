package main

import (
	"net/http"
)

func (app *application) internalError(w http.ResponseWriter, r *http.Request, err error) error {
	app.logger.Errorw("internal server error", "path", r.URL, "method", r.Method, "error", err)

	return app.writeJsonError(w, http.StatusInternalServerError, "something went wrong")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error, customErrorMsg string) error {
	app.logger.Warnw("bad request error", "path", r.URL, "method", r.Method, "error", err, "custom error message", customErrorMsg)
	if customErrorMsg == "" {
		customErrorMsg = err.Error()
	}
	return app.writeJsonError(w, http.StatusBadRequest, customErrorMsg)
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error, customErrorMsg string) error {
	app.logger.Warnw("bad request error", "path", r.URL, "method", r.Method, "error", err, "custom error message", customErrorMsg)
	if customErrorMsg == "" {
		customErrorMsg = err.Error()
	}
	return app.writeJsonError(w, http.StatusNotFound, customErrorMsg)
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err)
	app.writeJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) forbiddenRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("forbidden request error", "path", r.URL, "method", r.Method, "error", err)
	app.writeJsonError(w, http.StatusForbidden, "forbidden")
}
