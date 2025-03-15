package main

import (
	"net/http"

	"github.com/9thDuck/chat_go.git/internal/store"
)

const paginationCtxKey ctxKey = "pagination"

func getPaginationOptionsFromCtx(r *http.Request) *store.Pagination {
	return r.Context().Value(paginationCtxKey).(*store.Pagination)
}
