package film

import (
	"cinema/internal/lib/file"
	"cinema/internal/lib/sl"
	"cinema/internal/showcase/domain"
	"cinema/internal/showcase/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

var allowedPictureMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type UploadImageResult struct {
	PresignedURL string
	Key          string
}

type Film struct {
	log          *slog.Logger
	filmProvider Provider
	fileStorage  FileStorage
}

func New(
	log *slog.Logger,
	filmProvider Provider,
	fileStorage FileStorage,
) *Film {
	return &Film{log: log, filmProvider: filmProvider, fileStorage: fileStorage}
}

type Provider interface {
	Save(
		ctx context.Context,
		name string,
		description string,
	) (string, error)
	GetById(
		ctx context.Context,
		id string,
	) (domain.Film, error)
	UpdatePosterUrl(
		ctx context.Context,
		filmId string,
		posterUrl string,
	) error
}

type FileStorage interface {
	GetRestrictedPresignedUploadURL(ctx context.Context, key string, contentType string) (string, error)
	GetFileContentType(ctx context.Context, key string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	GetFilePath(key string) string
	MoveFile(ctx context.Context, keyFrom string, keyTo string) error
}

func (f *Film) Create(
	ctx context.Context,
	name string,
	description string,
) (string, error) {
	const op = "showcase.film.create"

	log := f.log.With(
		slog.String("op", op),
		slog.String("name", name),
		slog.String("description", description),
	)

	log.Info("creating new film")

	id, err := f.filmProvider.Save(
		ctx,
		name,
		description,
	)
	if err != nil {
		log.Error("failed to save film", sl.Err(err))

		return "", sl.WrapErr(op, err)
	}

	return id, nil
}

func (f *Film) Get(
	ctx context.Context,
	id string,
) (domain.Film, error) {
	const op = "showcase.film.get"

	log := f.log.With(
		slog.String("op", op),
		slog.String("id", id),
	)

	log.Info("getting film")

	film, err := f.filmProvider.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrFilmNotFound) {
			log.Warn("film not found", sl.Err(err))

			return domain.Film{}, sl.WrapErr(op, err)
		}

		log.Error("failed to get film", sl.Err(err))

		return domain.Film{}, sl.WrapErr(op, err)
	}

	if film.PosterUrl != nil {
		posterFullPath := f.fileStorage.GetFilePath(*film.PosterUrl)
		film.PosterUrl = &posterFullPath
	}

	return film, nil
}

func (f *Film) UploadImage(
	ctx context.Context,
	filmId string,
	contentType string,
) (UploadImageResult, error) {
	const op = "showcase.film.upload_image"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
		slog.String("contentType", contentType),
	)

	log.Info("adding image")

	ext, ok := allowedPictureMIMETypes[contentType]
	if !ok {
		log.Warn("not allowed file mime type", sl.Err(ErrIncorrectMIMEType))

		return UploadImageResult{}, sl.WrapErr(op, ErrIncorrectMIMEType)
	}

	_, err := f.filmProvider.GetById(ctx, filmId)
	if err != nil {
		if errors.Is(err, storage.ErrFilmNotFound) {
			log.Warn("film not found", sl.Err(err))

			return UploadImageResult{}, sl.WrapErr(op, ErrFilmNotFound)
		}

		log.Error("failed to get film", sl.Err(err))

		return UploadImageResult{}, sl.WrapErr(op, err)
	}

	fileName := uuid.New().String() + ext

	key := fmt.Sprintf("tmp/%s/%s", filmId, fileName)

	presignedUrl, err := f.fileStorage.GetRestrictedPresignedUploadURL(ctx, key, contentType)
	if err != nil {
		log.Error("failed to get presigned url", sl.Err(err))

		return UploadImageResult{}, sl.WrapErr(op, err)
	}

	return UploadImageResult{PresignedURL: presignedUrl, Key: key}, nil
}

func (f *Film) UpdatePoster(
	ctx context.Context,
	filmId string,
	key string,
) error {
	const op = "showcase.film.update_poster"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
		slog.String("key", key),
	)

	log.Info("updating poster")

	contentType, err := f.fileStorage.GetFileContentType(ctx, key)
	if err != nil {
		if errors.Is(err, file.ErrFileNotFound) {
			log.Warn("file not found", sl.Err(err))

			return sl.WrapErr(op, ErrFileNotFound)
		}

		log.Error("failed to get file content-type", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	ext, ok := allowedPictureMIMETypes[contentType]
	if !ok {
		log.Warn("not allowed file mime type", sl.Err(ErrIncorrectMIMEType))

		err = f.fileStorage.DeleteFile(ctx, key)
		if err != nil {
			log.Error("failed to delete file", sl.Err(err))
		}

		return sl.WrapErr(op, ErrIncorrectMIMEType)
	}

	film, err := f.filmProvider.GetById(ctx, filmId)
	if err != nil {
		if errors.Is(err, storage.ErrFilmNotFound) {
			log.Warn("film not found", sl.Err(err))

			return sl.WrapErr(op, err)
		}
		log.Error("failed to get film", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	fileName := uuid.New().String() + ext
	newKey := "films/" + filmId + "/" + fileName

	err = f.fileStorage.MoveFile(ctx, key, newKey)
	if err != nil {
		log.Error("failed to move file", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	err = f.filmProvider.UpdatePosterUrl(ctx, filmId, newKey)
	if err != nil {
		log.Error("failed to update poster", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	if film.PosterUrl != nil {
		err = f.fileStorage.DeleteFile(ctx, *film.PosterUrl)
		if err != nil {
			log.Error("failed to delete old poster file", sl.Err(err))
		}
	}

	return nil
}
