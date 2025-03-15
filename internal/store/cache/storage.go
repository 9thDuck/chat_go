package cache

import (
	"context"

	"github.com/9thDuck/chat_go.git/internal/store"
)

type Storage struct {
	Users interface {
		Get(ctx context.Context, userID int64) (*store.User, error)
		Set(ctx context.Context, user *store.User) error
		Delete(ctx context.Context, userID int64) error
	}
	Contacts interface {
		GetContactExists(ctx context.Context, userID, contactID int64) (bool, error)
		SetContactExists(ctx context.Context, userID, contactID int64, exists bool) error
		DeleteContactExists(ctx context.Context, userID, contactID int64) error
	}
	Misc interface {
		Close() error
	}
}
