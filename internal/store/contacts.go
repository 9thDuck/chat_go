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

func (s *ContactsStore) Get(ctx context.Context, userID int64, pagination *Pagination) (*[]Contact, int, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		SELECT user_id, contact_id, created_at, updated_at, COUNT(*) OVER() AS total
		FROM contacts
		WHERE user_id = $1 OR contact_id = $1
		ORDER BY ` + pagination.Sort + ` ` + pagination.SortDirection + `
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, userID, pagination.Limit, pagination.CalculateOffset())
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	contacts := make([]Contact, 0, pagination.Limit)
	total := 0

	for rows.Next() {
		contact := Contact{}
		err := rows.Scan(&contact.UserID, &contact.ContactID, &contact.CreatedAt, &contact.UpdatedAt, &total)
		if err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, contact)
	}

	return &contacts, total, nil
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
