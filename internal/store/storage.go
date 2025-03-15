package store

import (
	"context"
	"database/sql"
	"time"
)

type Storage struct {
	Users interface {
		Create(ctx context.Context, userP *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByID(ctx context.Context, userP *User) error
		UpdateUserDataByID(ctx context.Context, user *User) error
	}

	Roles interface {
		GetByName(ctx context.Context, roleName string) (*Role, error)
	}

	Contacts interface {
		Get(ctx context.Context, userID int64, pagination *Pagination) (*[]Contact, int, error)
		GetContactExists(ctx context.Context, userID, contactID int64) (bool, error)
		Delete(ctx context.Context, userID, contactID int64) error
	}

	ContactRequests interface {
		Create(ctx context.Context, senderID, receiverID int64) error
		Get(ctx context.Context, senderID int64, pagination *Pagination) (*[]ContactRequest, int, error)
		Accept(ctx context.Context, senderID, receiverID int64) error
		Reject(ctx context.Context, senderID, receiverID int64) error
		Delete(ctx context.Context, senderID, receiverID int64) error
	}

	Messages interface {
		Get(ctx context.Context, userID int64, pagination *Pagination) (*[]Message, int, error)
		Create(ctx context.Context, message *Message) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users:           &UsersStore{db},
		Roles:           &RolesStore{db},
		Contacts:        &ContactsStore{db},
		ContactRequests: &ContactRequestsStore{db},
		Messages:        &MessagesStore{db},
	}
}

var QueryTimeout = time.Second * 5

type WithTxFn func(*sql.Tx) error

func withTx(ctx context.Context, db *sql.DB, fn WithTxFn) error {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
