package postgres

import (
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (string, error) {
	query := `INSERT INTO sso.users (email, password_hash) VALUES ($1, $2) returning id`

	isUserExists, err := s.isUserExists(ctx, email)
	if err != nil {
		return "", err
	}

	if isUserExists {
		return "", storage.ErrUserExists
	}

	var id string
	err = s.pool.QueryRow(ctx, query, email, passHash).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT u.id, email, r.name, password_hash, created_at
		FROM sso.users u
		JOIN sso.roles r ON r.id = u.role_id
		WHERE email = $1
	`

	var user domain.User
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Role,
		&user.PassHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, storage.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *Storage) ChangeUserRole(ctx context.Context, userId string, role string) error {
	query := `
		UPDATE sso.users u
		SET role_id = (SELECT id FROM sso.roles WHERE name = $1)
		WHERE u.id = $2
	`

	result, err := s.pool.Exec(ctx, query, role, userId)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return storage.ErrUserNotFound
	}

	return err
}

func (s *Storage) isUserExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM sso.users WHERE email = $1)",
		email,
	).Scan(&exists)
	return exists, err
}
