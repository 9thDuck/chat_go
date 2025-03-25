package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/9thDuck/chat_go.git/internal/store"
)

const (
	userCtxKey            ctxKey = "user"
	userIDCtxKey          ctxKey = "userID"
	encryptionKeyIDCtxKey ctxKey = "encryptionKeyID"
)

type UpdateUserPayload struct {
	FirstName  string `json:"firstName" validate:"min=0,max=30"`
	LastName   string `json:"lastName" validate:"min=0,max=30"`
	ProfilePic string `json:"profilePic" validate:"min=0,max=255"`
}

// going to be unused
func (app *application) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	userIDFromParam := getUserIDParamFromCtx(r)

	app.logger.Infow("User ID from context", "otherUserID", userIDFromParam)

	if userIDFromParam == 0 {
		app.badRequestError(w, r, errors.New("user ID is required"), "")
		return
	}

	ctx := r.Context()
	user, err := app.getUser(ctx, userIDFromParam)
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

func (app *application) searchUsersByUsernamesAndGetIDsHandler(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	switch {
	case strings.TrimSpace(searchTerm) == "":
		app.badRequestError(w, r, errors.New("search term is required"), "")
		return
	case len(searchTerm) < 3:
		app.badRequestError(w, r, errors.New("search term must be at least 3 characters"), "")
		return
	case len(searchTerm) > 50:
		app.badRequestError(w, r, errors.New("search term must be less than 50 characters"), "")
		return
	}

	pagination := getPaginationOptionsFromCtx(r)
	userDataForAddContactSlice, total, err := app.store.Users.Search(r.Context(), getUserFromCtx(r).ID, searchTerm, pagination)

	switch err {
	case nil:
		if err := app.jsonResponse(w, http.StatusOK, paginatedEnvelope{Records: &userDataForAddContactSlice, TotalRecords: total}); err != nil {
			app.internalError(w, r, err)
			return
		}
	default:
		app.internalError(w, r, err)
		return
	}
}

func (app *application) getAuthenticatedUserHandler(w http.ResponseWriter, r *http.Request) {
	encryptionKeyID := getEncryptionKeyIDFromCtx(r)
	if encryptionKeyID == "" {
		app.badRequestError(w, r, errors.New("encryption key ID is required"), "")
		return
	}

	user, err := app.getUserWithEncryptionKey(r.Context(), getUserFromCtx(r).ID, encryptionKeyID)
	if err != nil {
		app.internalError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, &user); err != nil {
		app.internalError(w, r, err)
		return
	}
}
func (app *application) updateUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	var payload UpdateUserPayload
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err, "")
		return
	}

	if err := Validate.Struct(&payload); err != nil {
		app.badRequestError(w, r, err, "")
		return
	}

	user := getUserFromCtx(r)

	user.FirstName = payload.FirstName
	user.LastName = payload.LastName
	user.ProfilePic = payload.ProfilePic

	ctx := r.Context()

	if err := app.store.Users.UpdateUserDataByID(ctx, user); err != nil {
		app.internalError(w, r, err)
		return
	}

	if app.config.cacheCfg.initialised {
		if err := app.cache.Users.Delete(ctx, user.ID); err != nil {
			app.internalError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, &user); err != nil {
		app.internalError(w, r, err)
		return
	}
}

func getUserIDParamFromCtx(r *http.Request) int64 {
	val := r.Context().Value(userIDCtxKey)
	if val == nil {
		return 0
	}

	id, ok := val.(int64)
	if !ok {
		return 0
	}

	return id
}

func getUserFromCtx(r *http.Request) *store.User {
	data := r.Context().Value(userCtxKey)
	user, ok := data.(*store.User)
	if !ok {
		return nil
	}
	return user
}

func (app *application) deleteCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func getEncryptionKeyIDFromCtx(r *http.Request) string {
	id := r.Context().Value(encryptionKeyIDCtxKey)
	if id == nil {
		return ""
	}

	idStr, ok := id.(string)
	if !ok {
		return ""
	}
	return idStr
}
