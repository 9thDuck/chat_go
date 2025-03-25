package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type EncryptionKeysStore struct {
	db *sql.DB
}

type EncryptionKey struct {
	ID  string
	Key string
}

func (s *EncryptionKeysStore) Get(ctx context.Context, userID int64, encryptionKeyID string) (*EncryptionKey, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	SELECT key FROM encryption_keys
	WHERE user_id = $1 AND key_id = $2`

	var encryptionKey EncryptionKey
	err := s.db.QueryRowContext(ctx, query, userID, encryptionKeyID).Scan(&encryptionKey.Key)
	if err != nil {
		fmt.Println(err, "err")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &encryptionKey, nil
}

func (s *EncryptionKeysStore) Set(ctx context.Context, userID int64, encryptionKey *EncryptionKey) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	INSERT INTO encryption_keys (key_id, key, user_id)
	VALUES ($1, $2, $3)
	`
	_, err := s.db.ExecContext(ctx, query, encryptionKey.ID, encryptionKey.Key, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *EncryptionKeysStore) Delete(ctx context.Context, userID int64, encryptionKeyID string) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	DELETE FROM encryption_keys WHERE user_id = $1 AND key_id = $2`
	_, err := s.db.ExecContext(ctx, query, userID, encryptionKeyID)
	if err != nil {
		return err
	}
	return nil
}
