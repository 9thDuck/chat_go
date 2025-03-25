package cache

import (
	"context"
	"time"

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

	EncryptionKeys interface {
		Get(ctx context.Context, userID int64, encryptionKeyID string) (*store.EncryptionKey, error)
		Set(ctx context.Context, userID int64, encryptionKey *store.EncryptionKey) error
		Delete(ctx context.Context, userID int64, encryptionKeyID string) error
	}
}
type ExpiryTimes struct {
	Users          time.Duration
	Contacts       time.Duration
	EncryptionKeys time.Duration
}
