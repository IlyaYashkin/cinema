package postgres

import (
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/sl"
	"cinema/internal/showcase/domain"
	"cinema/internal/showcase/storage"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Film struct {
	*postgres.Postgres
}

func (f *Film) Save(
	ctx context.Context,
	name string,
	description string,
) (string, error) {
	const op = "showcase.storage.film.save"

	query := `INSERT INTO showcase.films(name, description) VALUES ($1, $2) RETURNING id`

	var id string
	err := f.Pool().QueryRow(ctx, query, name, description).Scan(&id)
	if err != nil {
		return "", sl.WrapErr(op, err)
	}

	return id, nil
}

func (f *Film) GetById(
	ctx context.Context,
	id string,
) (domain.Film, error) {
	const op = "showcase.storage.film.get_by_id"

	query := `SELECT id, name, description, poster_url FROM showcase.films WHERE id = $1`

	var film domain.Film
	err := f.Pool().QueryRow(ctx, query, id).Scan(
		&film.Id,
		&film.Name,
		&film.Description,
		&film.PosterUrl,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Film{}, sl.WrapErr(op, storage.ErrFilmNotFound)
	}
	if err != nil {
		return domain.Film{}, sl.WrapErr(op, err)
	}

	return film, nil
}

func (f *Film) UpdatePosterUrl(
	ctx context.Context,
	filmId string,
	posterUrl string,
) error {
	const op = "showcase.storage.film.update_poster_url"

	query := `UPDATE showcase.films SET poster_url = $1 WHERE id = $2`

	_, err := f.Pool().Exec(ctx, query, posterUrl, filmId)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}
