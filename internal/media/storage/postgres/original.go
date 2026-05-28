package postgres

import (
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/sl"
	"cinema/internal/media/domain"
	"cinema/internal/media/storage"
	"context"
	"database/sql"
	"errors"
)

type Original struct {
	*postgres.Postgres
}

func (o *Original) Create(
	ctx context.Context,
	filmId string,
	key string,
	status string,
) (string, error) {
	const op = "media.storage.original.create"

	query := `
		INSERT INTO media.originals (film_id, key, status) VALUES ($1, $2, $3)
		ON CONFLICT(key) DO NOTHING
		RETURNING id;
	`

	var id string
	err := o.Pool().QueryRow(ctx, query, filmId, key, status).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sl.WrapErr(op, storage.ErrOriginalKeyExists)
	}
	if err != nil {
		return "", sl.WrapErr(op, err)
	}

	return id, nil
}

func (o *Original) GetByKey(ctx context.Context, key string) (domain.Original, error) {
	const op = "media.storage.original.get_by_key"

	query := `SELECT id, film_id, key, status, created_at, updated_at FROM media.originals WHERE key = $1;`

	var original domain.Original
	err := o.Pool().QueryRow(ctx, query, key).Scan(
		&original.Id,
		&original.FilmId,
		&original.Key,
		&original.Status,
		&original.CreatedAt,
		&original.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Original{}, sl.WrapErr(op, storage.ErrOriginalNotFound)
	}
	if err != nil {
		return domain.Original{}, sl.WrapErr(op, err)
	}

	return original, nil
}

func (o *Original) DeleteByKey(ctx context.Context, key string) error {
	const op = "media.storage.original.delete_by_key"

	query := `DELETE FROM media.originals WHERE key = $1;`

	_, err := o.Pool().Exec(ctx, query, key)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (o *Original) UpdateStatus(ctx context.Context, id, status string) error {
	const op = "media.storage.original.update_status"

	query := `UPDATE media.originals SET status = $2 WHERE id = $1;`

	_, err := o.Pool().Exec(ctx, query, id, status)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}
