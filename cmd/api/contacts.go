package main

import (
	"net/http"

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
