package postgres

import (
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/sl"
	"cinema/internal/showcase/domain"
	"cinema/internal/showcase/storage"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Film struct {
	*postgres.Postgres
}

func (f *Film) Create(
	ctx context.Context,
	name string,
	description string,
) (string, error) {
	const op = "showcase.storage.film.create"

	query := `INSERT INTO showcase.films(name, description) VALUES ($1, $2) RETURNING id`

	var id string
	err := f.Pool().QueryRow(ctx, query, name, description).Scan(&id)
	if err != nil {
		return "", sl.WrapErr(op, err)
	}

	return id, nil
}

func (f *Film) Update(
	ctx context.Context,
	filmId string,
	name string,
	description string,
) error {
	const op = "showcase.storage.film.update"

	query := `UPDATE showcase.films SET name = $1, description = $2 WHERE id = $3`

	result, err := f.Pool().Exec(ctx, query, name, description, filmId)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrFilmNotFound)
	}

	return nil
}

func (f *Film) Delete(
	ctx context.Context,
	id string,
) error {
	const op = "showcase.storage.film.delete"

	query := `DELETE FROM showcase.films WHERE id = $1`

	result, err := f.Pool().Exec(ctx, query, id)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrFilmNotFound)
	}

	return nil
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

	images, err := f.GetImages(ctx, id)
	if err != nil {
		return domain.Film{}, sl.WrapErr(op, err)
	}

	film.Images = images

	return film, nil
}

func (f *Film) UpdatePoster(
	ctx context.Context,
	filmId string,
	posterUrl string,
) error {
	const op = "showcase.storage.film.update_poster"

	query := `UPDATE showcase.films SET poster_url = $1 WHERE id = $2`

	result, err := f.Pool().Exec(ctx, query, posterUrl, filmId)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrFilmNotFound)
	}

	return nil
}

func (f *Film) DeletePoster(
	ctx context.Context,
	filmId string,
) error {
	const op = "showcase.storage.film.delete_poster"

	query := `UPDATE showcase.films SET poster_url = NULL WHERE id = $1`

	result, err := f.Pool().Exec(ctx, query, filmId)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if result.RowsAffected() == 0 {
		return sl.WrapErr(op, storage.ErrFilmNotFound)
	}

	return nil
}

func (f *Film) UpdateImages(
	ctx context.Context,
	filmId string,
	images []string,
) (int64, error) {
	const op = "showcase.storage.film.update_images"

	var rows [][]any
	for _, image := range images {
		rows = append(rows, []any{filmId, image})
	}

	count, err := f.Pool().CopyFrom(
		ctx,
		pgx.Identifier{"showcase", "film_images"},
		[]string{"film_id", "url"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, sl.WrapErr(op, err)
	}

	return count, nil
}

func (f *Film) DeleteImages(
	ctx context.Context,
	filmId string,
	imageIds []int64,
) error {
	const op = "showcase.storage.film.delete_images"

	placeholders := make([]string, len(imageIds))
	args := make([]any, len(imageIds)+1)
	args[0] = filmId
	for i, id := range imageIds {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := "DELETE FROM showcase.film_images WHERE film_id = $1 AND id IN (" + strings.Join(placeholders, ",") + ")"

	_, err := f.Pool().Exec(ctx, query, args...)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (f *Film) GetImages(
	ctx context.Context,
	filmId string,
) ([]domain.FilmImage, error) {
	const op = "showcase.storage.film.get_images"

	query := `SELECT id, url FROM showcase.film_images WHERE film_id = $1`

	rows, err := f.Pool().Query(ctx, query, filmId)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	images, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.FilmImage, error) {
		var filmImage domain.FilmImage
		return filmImage, row.Scan(&filmImage.Id, &filmImage.Url)
	})
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	return images, nil
}
