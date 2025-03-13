package main

import (
	"encoding/json"
	"errors"
	"io"
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

func readJson(w http.ResponseWriter, r *http.Request, target any) error {
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		if err == io.EOF {
			return errors.New("request body is required")
		}
		return err
	}
	return nil
}
