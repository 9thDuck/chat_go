package store

import (
	"context"
	"database/sql"

	"github.com/9thDuck/chat_go.git/internal/domain"
)

type RolesStore struct {
	db *sql.DB
}

func (s *RolesStore) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
	SELECT
	id, level, description
	FROM roles
	WHERE name=$1`

	role := domain.Role{Name: name}

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
