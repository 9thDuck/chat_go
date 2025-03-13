package store

import (
	"context"
	"database/sql"
	"time"
)

type Storage struct {
	Users interface {
		Create(ctx context.Context, userP *User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{Users: &UserStorage{db}}
}

var QueryTimeout = time.Second * 5
