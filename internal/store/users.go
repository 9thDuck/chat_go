package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage struct {
	db *sql.DB
}

type User struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	HashedPassword string `json:"-"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func NewUser(username, email, firstName, lastName string) (user *User) {
	return &User{
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
}

func (u *User) SetHashedPassword(password string) error {
	hashByteArr, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashByteArr)
	return nil
}

func (u *User) ValidateCredentials(password string) bool {
	return bcrypt.
		CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)) == nil
}

func (s *UserStorage) Create(ctx context.Context, userP *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query :=
		`INSERT INTO users (
			username, email, hashed_password, first_name, last_name
		) VALUES (
		 	$1, $2, $3, $4, $5
		) RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		userP.Username,
		userP.Email,
		userP.HashedPassword,
		userP.FirstName,
		userP.LastName,
	).Scan(
		&userP.ID,
		&userP.CreatedAt,
		&userP.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == PQ_CODE_UNIQUE_CONSTRAINT_VIOLATION {
				return ErrConflict
			}
		}
		return err
	}

	return nil
}

func (s *UserStorage) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query :=
		`SELECT
		 id, username, hashed_password, first_name, last_name, created_at, updated_at 
		 FROM users 
		 WHERE email = $1`

	user := User{
		Email: email,
	}

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.HashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
