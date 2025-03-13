package main

import (
	"errors"
	"net/http"

	"github.com/9thDuck/chat_go.git/internal/store"
)

type SignupPayload struct {
	Username  string `json:"username" validate:"required,min=8,max=30"`
	Email     string `json:"email" validate:"email,required,max=150"`
	Password  string `json:"password" validate:"required,min=8,max=20"`
	FirstName string `json:"first_name" validate:"omitempty,min=8,max=30"`
	LastName  string `json:"last_name" validate:"omitempty,min=8,max=30"`
}

func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {
	payload := SignupPayload{}
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err, "")
		return
	}
	if err := Validate.Struct(&payload); err != nil {
		app.badRequestError(w, r, err, "")
		return
	}

	ctx := r.Context()

	userP := store.NewUser(
		payload.Username,
		payload.Email,
		payload.FirstName,
		payload.LastName,
	)
	userP.SetHashedPassword(payload.Password)

	if err := app.store.Users.Create(ctx, userP); err != nil {
		if errors.Is(err, store.ErrConflict) {
			app.badRequestError(w, r, err, "")
			return
		}
		app.internalError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, userP); err != nil {
		app.internalError(w, r, err)
		return
	}
}
