package postgres

import (
	"cinema/internal/lib/postgres"
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
	query := `INSERT INTO showcase.films(name, description) VALUES ($1, $2) RETURNING id`

	var id string
	err := f.Pool().QueryRow(ctx, query, name, description).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (f *Film) GetById(
	ctx context.Context,
	id string,
) (domain.Film, error) {
	query := `SELECT id, name, description, poster_url FROM showcase.films WHERE id = $1`

	var film domain.Film
	err := f.Pool().QueryRow(ctx, query, id).Scan(
		&film.Id,
		&film.Name,
		&film.Description,
		&film.PosterUrl,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Film{}, storage.ErrFilmNotFound
	}
	if err != nil {
		return domain.Film{}, err
	}

	return film, nil
}
