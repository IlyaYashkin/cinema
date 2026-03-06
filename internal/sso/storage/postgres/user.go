package postgres

import (
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (string, error) {
	sql := `insert into sso.users (email, password_hash) values ($1, $2) returning id`

	isUserExists, err := s.isUserExists(ctx, email)
	if err != nil {
		return "", err
	}

	if isUserExists {
		return "", storage.ErrUserExists
	}

	var id string
	err = s.pool.QueryRow(ctx, sql, email, passHash).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	sql := `select id, email, password_hash, created_at from sso.users where email = $1`

	var user domain.User
	err := s.pool.QueryRow(ctx, sql, email).Scan(
		&user.Id,
		&user.Email,
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

func (s *Storage) isUserExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		"select exists(select 1 from sso.users where email = $1)",
		email,
	).Scan(&exists)
	return exists, err
}
