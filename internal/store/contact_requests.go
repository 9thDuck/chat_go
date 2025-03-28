package store

import (
	"context"
	"database/sql"

	"github.com/9thDuck/chat_go.git/internal/domain"
	"github.com/lib/pq"
)

type ContactRequestsStore struct {
	db *sql.DB
}

type ContactRequest domain.ContactRequest

func (s *ContactRequestsStore) Create(ctx context.Context, senderID, receiverID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	WITH existing_contacts AS (
		SELECT 1
		FROM contacts
		WHERE (user_id = $1 AND contact_id = $2) OR (user_id = $2 AND contact_id = $1)
	),
	existing_requests AS (
		SELECT status
		FROM contact_requests
		WHERE (sender_id = $1 AND receiver_id = $2)
		ORDER BY created_at DESC
		LIMIT 1
	)
	INSERT INTO contact_requests (sender_id, receiver_id, status)
	SELECT $1, $2, 'pending'
	WHERE NOT EXISTS (SELECT 1 FROM existing_contacts)
	AND NOT EXISTS (SELECT 1 FROM existing_requests WHERE status != 'rejected')
	RETURNING sender_id;`

	var returnedSenderID int64
	err := s.db.QueryRowContext(ctx, query, senderID, receiverID).Scan(&returnedSenderID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrContactRequestAlreadyExists
		}
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case PQ_CODE_UNIQUE_CONSTRAINT_VIOLATION:
				return ErrContactRequestAlreadyExists
			case PQ_CODE_FOREIGN_KEY_CONSTRAINT_VIOLATION:
				return ErrContactRequestForeignKeyViolation
			default:
				return err
			}
		}
		return err
	}

	return nil
}

func (s *ContactRequestsStore) Get(ctx context.Context, senderID int64, pagination *Pagination) (*[]ContactRequest, int, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		SELECT c.sender_id, c.receiver_id, c.created_at, u.username AS sender_username, 
		u2.username AS receiver_username, m.content, COUNT(*) OVER() AS total
		FROM contact_requests c 
		JOIN users u ON c.sender_id = u.id
		JOIN users u2 ON c.receiver_id = u2.id
		JOIN messages m ON m.sender_id = c.sender_id 
		  AND m.receiver_id = c.receiver_id
		  AND m.created_at BETWEEN c.created_at - interval '1 second' AND c.created_at + interval '1 second'
		WHERE (c.sender_id = $1 OR c.receiver_id = $1) AND c.status = 'pending'
		ORDER BY ` + pagination.Sort + ` ` + pagination.SortDirection + `
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, senderID, pagination.Limit, pagination.CalculateOffset())
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	contactRequests := make([]ContactRequest, 0, pagination.Limit)
	total := 0

	for rows.Next() {
		contactRequest := ContactRequest{}
		err := rows.Scan(
			&contactRequest.SenderID,
			&contactRequest.ReceiverID,
			&contactRequest.CreatedAt,
			&contactRequest.SenderUsername,
			&contactRequest.ReceiverUsername,
			&contactRequest.MessageContent,
			&total,
		)
		if err != nil {
			return nil, 0, err
		}
		contactRequests = append(contactRequests, contactRequest)
	}

	return &contactRequests, total, nil
}

func acceptContactRequest(ctx context.Context, tx *sql.Tx, senderID, receiverID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		UPDATE contact_requests
		SET status = $3
		WHERE sender_id = $1 AND receiver_id = $2`

	res, err := tx.ExecContext(ctx, query, senderID, receiverID, "accepted")
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrContactRequestNotFound
	}

	return nil
}

func (s *ContactRequestsStore) Accept(ctx context.Context, senderID, receiverID int64) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		err := acceptContactRequest(ctx, tx, senderID, receiverID)
		if err != nil {
			return err
		}

		err = createContact(ctx, tx, senderID, receiverID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *ContactRequestsStore) Reject(ctx context.Context, senderID, receiverID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		UPDATE contact_requests
		SET status = $3
		WHERE sender_id = $1 AND receiver_id = $2`

	res, err := s.db.ExecContext(ctx, query, senderID, receiverID, "rejected")
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrContactRequestNotFound
	}

	return nil
}

func (s *ContactRequestsStore) Delete(ctx context.Context, senderID, receiverID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		DELETE FROM contact_requests
		WHERE sender_id = $1 AND receiver_id = $2 AND status = 'pending'
		RETURNING sender_id;`

	var returnedSenderID int64
	err := s.db.QueryRowContext(ctx, query, senderID, receiverID).Scan(&returnedSenderID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrContactRequestNotFound
		default:
			return err
		}
	}

	return nil
}
