package store

import (
	"context"
	"database/sql"
)

type RolesStore struct {
	db *sql.DB
}

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

func (s *RolesStore) GetByName(ctx context.Context, name string) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	SELECT
	id, level, description
	FROM roles
	WHERE name=$1`

	role := Role{Name: name}

	err := s.db.QueryRowContext(
		ctx,
		query,
		name).
		Scan(
			&role.ID,
			&role.Level,
			&role.Description,
		)

	if err != nil {
		return nil, err
	}

	return &role, nil
}
