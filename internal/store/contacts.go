package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type ContactsStore struct {
	db *sql.DB
}

type Contact struct {
	UserID    int64  `json:"user_id"`
	ContactID int64  `json:"contact_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *ContactsStore) Get(ctx context.Context, userID int64, pagination *Pagination) (*[]int64, int, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	query := `
		SELECT (
			CASE 
				WHEN user_id = $1 THEN contact_id
				ELSE user_id
			END
		), COUNT(*) OVER() AS total
		FROM contacts
		WHERE user_id = $1 OR contact_id = $1
		LIMIT $2 OFFSET $3`


	rows, err := s.db.QueryContext(ctx, query, userID, pagination.Limit, pagination.CalculateOffset())
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

		contactIDSlice := make([]int64, 0, pagination.Limit)
	total := 0

	for rows.Next() {
			var contactID int64
			err := rows.Scan(&contactID, &total)
		if err != nil {
			return nil, 0, err
		}
			contactIDSlice = append(contactIDSlice, contactID)
	}

	return &contactIDSlice, total, nil
}

func (s *ContactsStore) GetContactExists(ctx context.Context, userID, contactID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1 FROM contacts 
			WHERE (user_id = $1 AND contact_id = $2) 
			OR (user_id = $2 AND contact_id = $1)
		)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID, contactID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *ContactsStore) Delete(ctx context.Context, userID, contactID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		DELETE FROM contacts
		WHERE (user_id = $1 AND contact_id = $2) OR (user_id = $2 AND contact_id = $1)`

	_, err := s.db.ExecContext(ctx, query, userID, contactID)
	if err != nil {
		return err
	}

	return nil
}

func createContact(ctx context.Context, tx *sql.Tx, userID, contactID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		INSERT INTO contacts (user_id, contact_id)
		VALUES ($1, $2)`

	_, err := tx.ExecContext(ctx, query, userID, contactID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == PQ_CODE_UNIQUE_CONSTRAINT_VIOLATION {
				return ErrContactAlreadyExists
			}
		}
		return err
	}

	return nil
}
