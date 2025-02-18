package store

import "database/sql"

type Storage struct {
	Users interface {
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{Users: &UserStorage{db}}
}
