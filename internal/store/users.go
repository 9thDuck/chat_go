package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UsersStore struct {
	db *sql.DB
}

type UserDataForAddContact struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	PublicKey       string `json:"public_key"`
	IsContact       bool   `json:"is_contact"`
	HasPendingRequest bool `json:"has_pending_request"`
}

type User struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	HashedPassword string `json:"-"`
	PublicKey      string `json:"public_key"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePic     string `json:"profile_pic"`
	RoleID         int64  `json:"role_id"`
	Role           *Role  `json:"role"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func NewUser(username, email, firstName, lastName, publicKey string) (user *User) {
	return &User{
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		PublicKey: publicKey,
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

func (s *UsersStore) Create(ctx context.Context, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	query := `WITH inserted_user AS (
			INSERT INTO users (
				username,
				hashed_password,
				email, 
				first_name,
				last_name,
				profile_pic,
				role_id,
				public_key
			)
			VALUES ($1, $2, $3, $4, $5, $6,
				(SELECT r.id FROM roles r WHERE r.name = $7),
				$8
			)
			RETURNING id, role_id, created_at, updated_at, public_key
		)
		SELECT 
			iu.id, 
			iu.role_id, 
			iu.role_id,
			r.level, 
			r.description,
			iu.created_at, 
			iu.updated_at,
			iu.public_key
		FROM inserted_user iu
		JOIN roles r ON r.id = iu.role_id;`

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.HashedPassword,
		user.Email,
		user.FirstName,
		user.LastName,
		user.ProfilePic,
		user.Role.Name,
		user.PublicKey,
	).Scan(
		&user.ID,
		&user.RoleID,
		&user.Role.ID,
		&user.Role.Level,
		&user.Role.Description,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.PublicKey,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			pqErrorMsg := pqErr.Error()
			switch {
			case strings.Contains(pqErrorMsg, "users_email_key"):
				return ErrDuplicateMail
			case strings.Contains(pqErrorMsg, "users_username_key"):
				return ErrDuplicateUsername
			default:
				return err
			}
		}
		return err
	}
	return nil
}

func (s *UsersStore) GetByID(ctx context.Context, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	query := `
		SELECT 
		u.username, u.email, u.hashed_password, u.first_name, u.last_name, u.public_key, u.role_id, u.created_at, u.updated_at,
		r.id, r.name, r.level, r.description
		FROM 
		users u JOIN roles r ON u.role_id = r.id
		WHERE u.id=$1`

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.ID,
	).Scan(
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.PublicKey,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *UsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query :=
		`SELECT
		 u.id, u.username, u.hashed_password, u.first_name, u.last_name, u.public_key, u.role_id, u.created_at, u.updated_at,
		 r.id, r.name, r.description, r.level
		 FROM users u JOIN roles r ON r.id = u.role_id
		 WHERE u.email = $1`

	user := User{
		Email: email,
		Role:  &Role{},
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
		&user.PublicKey,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.Level,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UsersStore) UpdateUserDataByID(ctx context.Context, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, profile_pic = $3 
		WHERE id = $4`

	_, err := s.db.ExecContext(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.ProfilePic,
		user.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersStore) Search(ctx context.Context, userID int64, searchTerm string, pagination *Pagination) (*[]UserDataForAddContact, int, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	
	query := `
		WITH contact_status AS (
			SELECT contact_id as id, true as is_contact
			FROM contacts
			WHERE user_id = $1
			UNION
			SELECT user_id as id, true as is_contact
			FROM contacts
			WHERE contact_id = $1
		),
		pending_requests AS (
			SELECT receiver_id as id, true as has_pending_request
			FROM contact_requests
			WHERE sender_id = $1 AND status = 'pending'
			UNION
			SELECT sender_id as id, true as has_pending_request
			FROM contact_requests
			WHERE receiver_id = $1 AND status = 'pending'
		)
		SELECT 
			u.id, 
			u.username, 
			u.public_key,
			COALESCE(cs.is_contact, false) as is_contact,
			COALESCE(pr.has_pending_request, false) as has_pending_request,
			COUNT(*) OVER() AS total_count
		FROM users u
		LEFT JOIN contact_status cs ON cs.id = u.id
		LEFT JOIN pending_requests pr ON pr.id = u.id
		WHERE u.username ILIKE $2
		AND u.id != $1
		ORDER BY 
			cs.is_contact DESC NULLS LAST,
			pr.has_pending_request DESC NULLS LAST,
			u.username ASC
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + searchTerm + "%"

	rows, err := s.db.QueryContext(
		ctx, 
		query, 
		userID, 
		searchPattern, 
		pagination.Limit, 
		pagination.CalculateOffset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	totalCount := 0
	userDataForAddContactSlice := make([]UserDataForAddContact, 0, pagination.Limit)
	for rows.Next() {
		var userDataForAddContact UserDataForAddContact
		if err := rows.Scan(
			&userDataForAddContact.ID,
			&userDataForAddContact.Username,
			&userDataForAddContact.PublicKey,
			&userDataForAddContact.IsContact,
			&userDataForAddContact.HasPendingRequest,
			&totalCount,
		); err != nil {
			return nil, 0, err
		}
		userDataForAddContactSlice = append(userDataForAddContactSlice, userDataForAddContact)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return &userDataForAddContactSlice, totalCount, nil
}
