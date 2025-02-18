package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func New(addr string, maxOpenConns, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	maxIdleDuration, err := time.ParseDuration(maxIdleTime)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(maxIdleDuration)

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
