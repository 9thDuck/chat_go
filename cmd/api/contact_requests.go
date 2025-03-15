package main

import (
	"net/http"

	"github.com/9thDuck/chat_go.git/internal/store"
)

const contactIDCtxKey ctxKey = "contactID"

func (app *application) getContactRequestByIDHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	pagination := getPaginationOptionsFromCtx(r)

	contactRequests, total, err := app.store.ContactRequests.Get(r.Context(), user.ID, pagination)
	if err != nil {
		app.internalError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, paginatedEnvelope{Records: &contactRequests, TotalRecords: total}); err != nil {
		app.internalError(w, r, err)
		return
	}
}

func (app *application) createContactRequestHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	contactID := getContactIDFromCtx(r)

	err := app.store.ContactRequests.Create(r.Context(), user.ID, contactID)
	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
		return
	case store.ErrContactRequestAlreadyExists:
		app.badRequestError(w, r, err, "")
		return
	case store.ErrContactRequestForeignKeyViolation:
		app.notFoundError(w, r, err, "")
		return
	default:
		app.internalError(w, r, err)
		return
	}

}

func (app *application) updateContactRequestHandler(w http.ResponseWriter, r *http.Request) {
	operation := r.URL.Query().Get("operation")

	if operation != "accept" && operation != "reject" {
		app.badRequestError(w, r, nil, "invalid operation. Operation can only be \"accept\" or \"reject\"")
		return
	}

	user := getUserFromCtx(r)
	contactID := getContactIDFromCtx(r)

	var err error
	switch operation {
	case "accept":
		err = app.store.ContactRequests.Accept(r.Context(), contactID, user.ID)
		if err == nil && app.config.cacheCfg.initialised {
			err = app.cache.Contacts.SetContactExists(r.Context(), user.ID, contactID, true)
			if err != nil {
				app.logger.Errorw("Failed to update contacts cache", "error", err)
			}
		}
	case "reject":
		err = app.store.ContactRequests.Reject(r.Context(), contactID, user.ID)
	}

	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
		return
	case store.ErrContactRequestNotFound:
		app.notFoundError(w, r, err, "contact request not found")
		return
	case store.ErrContactAlreadyExists:
		app.badRequestError(w, r, err, "")
		return
	default:
		app.internalError(w, r, err)
		return
	}
}

func (app *application) deleteContactRequestHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	contactID := getContactIDFromCtx(r)

	err := app.store.ContactRequests.Delete(r.Context(), user.ID, contactID)
	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	case store.ErrContactRequestNotFound:
		app.notFoundError(w, r, err, "")
		return
	default:
		app.internalError(w, r, err)
		return
	}
}

func getContactIDFromCtx(r *http.Request) int64 {
	return r.Context().Value(contactIDCtxKey).(int64)
}
