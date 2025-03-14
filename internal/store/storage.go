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
	}

	Roles interface {
		GetByName(ctx context.Context, roleName string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{Users: &UsersStore{db}}
}

var QueryTimeout = time.Second * 5
