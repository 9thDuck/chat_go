package main

import (
	"net/http"
)

func (app *application) getHomeHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.jsonResponse(w, http.StatusOK, "Hello"); err != nil {
		app.internalError(w, r, err)
	}
}
