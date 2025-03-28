package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/9thDuck/chat_go.git/internal/domain"
	"github.com/lib/pq"
)

type Message domain.Message

type MessageVersion struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"message_id"`
	Content   string `json:"content"`
	Version   int64  `json:"version"`
	CreatedAt string `json:"created_at"`
}

type MessagesStore struct {
	db *sql.DB
}

func (s *MessagesStore) Get(ctx context.Context, userID int64, pagination *Pagination) (*[]Message, int, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	SELECT 
		id,
		sender_id,
		receiver_id,
		content,
		is_read,
		is_delivered,
		version,
		edited,
		created_at,
		updated_at,
		COUNT(*) OVER() AS total
	FROM messages
	WHERE (sender_id = $1 OR receiver_id = $1) AND is_delivered = $2
	ORDER BY created_at ASC
	LIMIT $3 OFFSET $4`

	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
		false,
		pagination.Limit,
		pagination.CalculateOffset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	messages := make([]Message, 0, pagination.Limit)
	messageMap := make(map[int64]*Message)
	messageIDs := make([]int64, 0, len(messages))
	total := 0

	for rows.Next() {
		message := Message{}
		err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.IsRead,
			&message.IsDelivered,
			&message.Version,
			&message.Edited,
			&message.CreatedAt,
			&message.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, err
		}

		emptyAttachments := []string{}
		message.Attachments = &emptyAttachments

		messages = append(messages, message)
		if _, exists := messageMap[message.ID]; !exists {
			messageIDs = append(messageIDs, message.ID)
		}
		messageMap[message.ID] = &messages[len(messages)-1]
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(messages) > 0 {
		attachmentsQuery := `
			SELECT message_id, path
			FROM attachments
			WHERE message_id = ANY($1)
			ORDER BY message_id, id`

		attachmentRows, err := s.db.QueryContext(ctx, attachmentsQuery, pq.Array(messageIDs))
		if err != nil {
			return nil, 0, err
		}
		defer attachmentRows.Close()

		for attachmentRows.Next() {
			var messageID int64
			var path string

			if err := attachmentRows.Scan(&messageID, &path); err != nil {
				return nil, 0, err
			}

			if message, exists := messageMap[messageID]; exists {
				*message.Attachments = append(*message.Attachments, path)
			}
		}

		if err = attachmentRows.Err(); err != nil {
			return nil, 0, err
		}
	}

	return &messages, total, nil
}

func (s *MessagesStore) Create(ctx context.Context, message *Message) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		err := addMessage(ctx, tx, message)
		if err != nil {
			return err
		}

		if message.Attachments != nil && len(*message.Attachments) > 0 {
			err = addAttachments(ctx, tx, message.ID, message.Attachments)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *MessagesStore) Delete(ctx context.Context, messageID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM messages WHERE id = $1`, messageID)
	return err
}

func addAttachments(ctx context.Context, tx *sql.Tx, messageID int64, attachments *[]string) error {
	query := `
		INSERT INTO attachments 
		(message_id, path)
		SELECT $1, unnest($2::text[])`

	res, err := tx.ExecContext(ctx, query, messageID, pq.Array(*attachments))
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != int64(len(*attachments)) {
		return errors.New("failed to add all the attachments")
	}
	return nil
}

func addMessage(ctx context.Context, tx *sql.Tx, message *Message) error {
	query := `
	INSERT INTO messages 
		(sender_id, receiver_id, content, is_read, is_delivered, version, edited)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at, updated_at`

	err := tx.QueryRowContext(
		ctx,
		query,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		message.IsRead,
		message.IsDelivered,
		message.Version,
		message.Edited,
	).Scan(
		&message.ID,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
