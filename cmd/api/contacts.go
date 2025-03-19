package main

import (
	"net/http"
	"strings"

	"github.com/9thDuck/chat_go.git/internal/store"
)

func (app *application) getContactsHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	pagination := getPaginationOptionsFromCtx(r)

	contactIDsSlice, total, err := app.store.Contacts.Get(r.Context(), user.ID, pagination)
	if err != nil {
		app.internalError(w, r, err)
		return
	}

	contactsAsUsers := make([]store.User, len(*contactIDsSlice))
	for i, contactID := range *contactIDsSlice {
		user, err := app.getUser(r.Context(), contactID)
		if err != nil {
			app.internalError(w, r, err)
			return
		}
		contactsAsUsers[i] = *user
	}

	app.jsonResponse(w, http.StatusOK, paginatedEnvelope{Records: &contactsAsUsers, TotalRecords: total})
}

func (app *application) deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	contactID := getContactIDFromCtx(r)

	err := app.store.Contacts.Delete(r.Context(), user.ID, contactID)
	switch err {
	case nil:
		if app.config.cacheCfg.initialised {
			err = app.cache.Contacts.DeleteContactExists(r.Context(), user.ID, contactID)
			if err != nil {
				app.logger.Errorw("Failed to delete contact from cache", "error", err)
			}
		}
		w.WriteHeader(http.StatusNoContent)
		return
	case store.ErrContactNotFound:
		app.notFoundError(w, r, err, "contact not found")
		return
	default:
		app.internalError(w, r, err)
		return
	}
}

func (app *application) searchContactsHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	searchTerm := r.URL.Query().Get("q")

	switch {
	case strings.TrimSpace(searchTerm) == "":
		app.badRequestError(w, r, nil, "search term is required")
		return
	case len(searchTerm) < 3:
		app.badRequestError(w, r, nil, "search term must be at least 3 characters")
		return
	case len(searchTerm) > 50:
		app.badRequestError(w, r, nil, "search term must be less than 50 characters")
		return
	}

	pagination := getPaginationOptionsFromCtx(r)
	contactIDSlice, total, err := app.store.Contacts.Search(r.Context(), user.ID, searchTerm, pagination)
	if err != nil {
		app.internalError(w, r, err)
		return
	}
	contactsAsUsers := make([]store.User, len(*contactIDSlice))
	for i, contactID := range *contactIDSlice {
		user, err := app.getUser(r.Context(), contactID)
		if err != nil {
			app.internalError(w, r, err)
			return
		}
		contactsAsUsers[i] = *user
	}
	app.jsonResponse(w, http.StatusOK, paginatedEnvelope{Records: &contactsAsUsers, TotalRecords: total})
}
