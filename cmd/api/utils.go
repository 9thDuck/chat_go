package main

import (
	"encoding/json"
	"net/http"
)

type dataEnvelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error string `json:"error"`
}

func writeJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	return writeJson(w, status, &dataEnvelope{Data: data})
}

func (app *application) writeJsonError(w http.ResponseWriter, status int, errorMsg string) error {
	return writeJson(w, status, &errorEnvelope{Error: errorMsg})
}
