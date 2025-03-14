package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

const DefaultUserNotFoundErrMsg = "either credentials are invalid or user doesn't exist"

type SignupPayload struct {
	Username string `json:"username" validate:"required,min=8,max=30"`
	Email    string `json:"email" validate:"email,required,max=150"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type LoginPayload struct {
	Email    string `json:"email" validate:"email,required,max=150"`
	Password string `json:"password" validate:"required,min=8,max=20"`
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

	user := store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: &store.Role{
			Name: "user",
		},
	}

	user.SetHashedPassword(payload.Password)

	if err := app.store.Users.Create(ctx, &user); err != nil {
		switch err {
		case store.ErrDuplicateMail:
			app.badRequestError(w, r, err, "")
		case store.ErrDuplicateUsername:
			app.badRequestError(w, r, err, "")
		default:
			app.internalError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalError(w, r, err)
		return
	}
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err, "")
		return
	}

	if err := Validate.Struct(&payload); err != nil {
		app.badRequestError(w, r, err, "")
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)

	if err != nil {
		app.notFoundError(w, r, err, DefaultUserNotFoundErrMsg)
		return
	}

	if !user.ValidateCredentials(payload.Password) {
		app.badRequestError(w, r, errors.New(DefaultUserNotFoundErrMsg), "")
		return
	}

	accessTokenCookie, refreshTokenCookie, err := app.makeAuthCookiesSet(user.ID)
	if err != nil {
		app.internalError(w, r, err)
		return
	}

	http.SetCookie(w, accessTokenCookie)
	http.SetCookie(w, refreshTokenCookie)

	if err := app.jsonResponse(w, http.StatusOK, &user); err != nil {
		app.internalError(w, r, err)
		return
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	deleteCookie(w, "access_token")
	deleteCookie(w, "refresh_token")
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) makeAuthCookiesSet(userID int64) (accessCookie *http.Cookie, refreshCookie *http.Cookie, err error) {
	timeNow := time.Now()

	accessTokenClaims := jwt.MapClaims{
		"sub": userID,
		"iss": app.config.appName,
		"aud": app.config.appName,
		"exp": time.Now().Add(app.config.auth.token.exp.Access).Unix(),
		"nbf": timeNow.Unix(),
		"iat": timeNow.Unix(),
	}
	refreshTokenClaims := jwt.MapClaims{
		"sub": userID,
		"iss": app.config.appName,
		"aud": app.config.appName,
		"exp": time.Now().Add(app.config.auth.token.exp.Refresh).Unix(),
		"nbf": timeNow.Unix(),
		"iat": timeNow.Unix(),
	}

	accessToken, err := app.authenticator.GenerateToken(accessTokenClaims)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := app.authenticator.GenerateToken(refreshTokenClaims)
	if err != nil {
		return nil, nil, err
	}

	secure := app.config.env == "production"

	accessCookie = app.NewAuthCookie("access_token", accessToken, app.config.auth.token.exp.Access, secure)
	refreshCookie = app.NewAuthCookie("refresh_token", refreshToken, app.config.auth.token.exp.Refresh, secure)
	return accessCookie, refreshCookie, nil
}

func (app *application) NewAuthCookie(name, tokenString string, exp time.Duration, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    tokenString,
		Path:     "/",
		MaxAge:   int(exp),
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		HttpOnly: true,
	}
}
