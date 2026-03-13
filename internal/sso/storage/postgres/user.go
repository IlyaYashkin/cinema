package postgres

import (
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/sl"
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type User struct {
	*postgres.Postgres
}

func (u *User) SaveUser(ctx context.Context, email string, passHash []byte) (string, error) {
	const op = "sso.storage.user.save_user"

	query := `
		INSERT INTO sso.users (email, password_hash) VALUES ($1, $2)
		ON CONFLICT(email) DO NOTHING
		RETURNING id
	`

	var id string
	err := u.Pool().QueryRow(ctx, query, email, passHash).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", sl.WrapErr(op, storage.ErrUserExists)
		}

		return "", sl.WrapErr(op, err)
	}

	return id, nil
}

func (u *User) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	const op = "sso.storage.user.find_by_email"

	query := `
		SELECT u.id, email, r.name, password_hash, created_at
		FROM sso.users u
		JOIN sso.roles r ON r.id = u.role_id
		WHERE email = $1
	`

	var user domain.User
	err := u.Pool().QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Role,
		&user.PassHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, sl.WrapErr(op, storage.ErrUserNotFound)
	}
	if err != nil {
		return domain.User{}, sl.WrapErr(op, err)
	}

	return user, nil
}

func (u *User) FindById(ctx context.Context, id string) (domain.User, error) {
	const op = "sso.storage.user.find_by_id"

	query := `
		SELECT u.id, email, r.name, password_hash, created_at
		FROM sso.users u
		JOIN sso.roles r ON r.id = u.role_id
		WHERE u.id = $1
	`

	var user domain.User
	err := u.Pool().QueryRow(ctx, query, id).Scan(
		&user.Id,
		&user.Email,
		&user.Role,
		&user.PassHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, sl.WrapErr(op, storage.ErrUserNotFound)
	}
	if err != nil {
		return domain.User{}, sl.WrapErr(op, err)
	}

	return user, nil
}

func (u *User) UpdateUserRole(ctx context.Context, userId string, role string) error {
	const op = "sso.storage.user.update_user_role"

	query := `
		UPDATE sso.users u
		SET role_id = (SELECT id FROM sso.roles WHERE name = $1)
		WHERE u.id = $2
	`

	result, err := u.Pool().Exec(ctx, query, role, userId)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrUserNotFound)
	}

	return nil
}

func (u *User) UpdateUserEmail(ctx context.Context, userId string, email string) error {
	const op = "sso.storage.user.update_user_email"

	query := `
		UPDATE sso.users
		SET email = $2
		WHERE id = $1
	`

	result, err := u.Pool().Exec(ctx, query, userId, email)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrUserNotFound)
	}

	return nil
}

func (u *User) UpdateUserPassword(ctx context.Context, userId string, passHash []byte) error {
	const op = "sso.storage.user.update_user_password"

	query := `
		UPDATE sso.users
		SET password_hash = $2
		WHERE id = $1
	`

	result, err := u.Pool().Exec(ctx, query, userId, passHash)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrUserNotFound)
	}

	return nil
}
