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
	"strings"

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

type UpdateImagesResult struct {
	FailedKeys  []string
	UpdatedKeys []string
}

type DeleteImagesResult struct {
	FailedIds  []int64
	DeletedIds []int64
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
	Create(
		ctx context.Context,
		name string,
		description string,
	) (string, error)
	Update(
		ctx context.Context,
		filmId string,
		name string,
		description string,
	) error
	Delete(
		ctx context.Context,
		id string,
	) error
	GetById(
		ctx context.Context,
		id string,
	) (domain.Film, error)
	UpdatePoster(
		ctx context.Context,
		filmId string,
		posterUrl string,
	) error
	DeletePoster(
		ctx context.Context,
		filmId string,
	) error
	UpdateImages(
		ctx context.Context,
		filmId string,
		images []string,
	) (int64, error)
	DeleteImages(
		ctx context.Context,
		filmId string,
		imageIds []int64,
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

	id, err := f.filmProvider.Create(
		ctx,
		name,
		description,
	)
	if err != nil {
		log.Error("failed to create film", sl.Err(err))

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
	const op = "showcase.film.update"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
		slog.String("name", name),
		slog.String("description", description),
	)

	log.Info("updating film")

	err := f.filmProvider.Update(ctx, filmId, name, description)
	if errors.Is(err, storage.ErrFilmNotFound) {
		log.Warn("film not found")

		return sl.WrapErr(op, err)
	}
	if err != nil {
		log.Error("failed to update film", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	return nil
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

	for idx, img := range film.Images {
		imageFullPath := f.fileStorage.GetFilePath(img.Url)
		film.Images[idx].Url = imageFullPath
	}

	return film, nil
}

func (f *Film) Delete(
	ctx context.Context,
	filmId string,
) error {
	const op = "showcase.film.delete"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
	)

	log.Info("deleting film")

	film, err := f.filmProvider.GetById(ctx, filmId)
	if errors.Is(err, storage.ErrFilmNotFound) {
		log.Warn("film not found", sl.Err(err))

		return sl.WrapErr(op, ErrFilmNotFound)
	}
	if err != nil {
		log.Error("failed to get film", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	if film.PosterUrl != nil {
		err = f.fileStorage.DeleteFile(ctx, *film.PosterUrl)
		if err != nil {
			log.Error("failed to delete poster", sl.Err(err))

			return sl.WrapErr(op, err)
		}
	}

	for _, image := range film.Images {
		if err := f.fileStorage.DeleteFile(ctx, image.Url); err != nil {
			log.Error("failed to delete image", sl.Err(err))

			return sl.WrapErr(op, err)
		}
	}

	err = f.filmProvider.Delete(ctx, filmId)
	if err != nil {
		log.Error("failed to delete film", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	return nil
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

	err = f.filmProvider.UpdatePoster(ctx, filmId, newKey)
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

func (f *Film) DeletePoster(
	ctx context.Context,
	filmId string,
) error {
	const op = "showcase.film.delete_poster"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
	)

	log.Info("deleting poster")

	film, err := f.filmProvider.GetById(ctx, filmId)
	if errors.Is(err, storage.ErrFilmNotFound) {
		log.Warn("film not found", sl.Err(err))

		return sl.WrapErr(op, ErrFilmNotFound)
	}
	if err != nil {
		log.Error("failed to get film", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	if film.PosterUrl == nil {
		return nil
	}

	err = f.fileStorage.DeleteFile(ctx, *film.PosterUrl)
	if err != nil {
		log.Error("failed to delete poster file", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	err = f.filmProvider.DeletePoster(ctx, filmId)
	if err != nil {
		log.Error("failed to delete poster", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	return nil
}

func (f *Film) UpdateImages(
	ctx context.Context,
	filmId string,
	keys []string,
) (UpdateImagesResult, error) {
	const op = "showcase.film.update_images"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
		slog.String("keys", strings.Join(keys, ",")),
	)

	log.Info("updating images")

	_, err := f.filmProvider.GetById(ctx, filmId)
	if err != nil {
		if errors.Is(err, storage.ErrFilmNotFound) {
			log.Warn("film not found", sl.Err(err))

			return UpdateImagesResult{}, sl.WrapErr(op, err)
		}
		log.Error("failed to get film", sl.Err(err))

		return UpdateImagesResult{}, sl.WrapErr(op, err)
	}

	var failedKeys []string
	var newKeys []string

	for _, key := range keys {
		contentType, err := f.fileStorage.GetFileContentType(ctx, key)
		if err != nil {
			if errors.Is(err, file.ErrFileNotFound) {
				log.With(slog.String("key", key)).Warn("file not found", sl.Err(err))
			} else {
				log.With(slog.String("key", key)).Error("failed to get file content-type", sl.Err(err))
			}

			failedKeys = append(failedKeys, key)
			continue
		}

		ext, ok := allowedPictureMIMETypes[contentType]
		if !ok {
			log.With(slog.String("key", key)).Warn("not allowed file mime type", sl.Err(ErrIncorrectMIMEType))

			err = f.fileStorage.DeleteFile(ctx, key)
			if err != nil {
				log.With(slog.String("key", key)).Error("failed to delete file", sl.Err(err))
			}

			failedKeys = append(failedKeys, key)
			continue
		}

		fileName := uuid.New().String() + ext
		newKey := "films/" + filmId + "/" + fileName

		err = f.fileStorage.MoveFile(ctx, key, newKey)
		if err != nil {
			log.With(slog.String("key", key)).Error("failed to move file", sl.Err(err))

			failedKeys = append(failedKeys, key)
			continue
		}

		newKeys = append(newKeys, newKey)
	}

	if len(newKeys) > 0 {
		_, err = f.filmProvider.UpdateImages(ctx, filmId, newKeys)
		if err != nil {
			log.Error("failed to update images", sl.Err(err))

			return UpdateImagesResult{}, sl.WrapErr(op, err)
		}
	}

	return UpdateImagesResult{
		FailedKeys:  failedKeys,
		UpdatedKeys: newKeys,
	}, nil
}

func (f *Film) DeleteImages(
	ctx context.Context,
	filmId string,
	imageIds []int64,
) (DeleteImagesResult, error) {
	const op = "showcase.film.delete_images"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
		slog.String("imageIds", fmt.Sprintf("%v", imageIds)),
	)

	log.Info("deleting images")

	film, err := f.filmProvider.GetById(ctx, filmId)
	if errors.Is(err, storage.ErrFilmNotFound) {
		log.Warn("film not found", sl.Err(err))

		return DeleteImagesResult{}, sl.WrapErr(op, err)
	}
	if err != nil {
		log.Error("failed to get film", sl.Err(err))

		return DeleteImagesResult{}, sl.WrapErr(op, err)
	}

	filmImagesByIds := make(map[int64]domain.FilmImage, len(film.Images))
	for _, img := range film.Images {
		filmImagesByIds[img.Id] = img
	}

	var failedIds []int64
	var deletedIds []int64

	for _, id := range imageIds {
		filmImage, ok := filmImagesByIds[id]
		if !ok {
			failedIds = append(failedIds, id)
			continue
		}

		err = f.fileStorage.DeleteFile(ctx, filmImage.Url)
		if err != nil {
			log.Error("failed to delete file", sl.Err(err))

			failedIds = append(failedIds, id)
			continue
		}

		deletedIds = append(deletedIds, id)
	}

	if len(deletedIds) > 0 {
		err = f.filmProvider.DeleteImages(ctx, filmId, deletedIds)
		if err != nil {
			log.Error("failed to delete images", sl.Err(err))

			return DeleteImagesResult{}, sl.WrapErr(op, err)
		}
	}

	return DeleteImagesResult{
		FailedIds:  failedIds,
		DeletedIds: deletedIds,
	}, nil
}
