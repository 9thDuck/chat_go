package store

import "database/sql"

type UserStorage struct {
	db *sql.DB
}
