package film

import (
	"cinema/internal/lib/sl"
	"cinema/internal/showcase/domain"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

var allowedPictureMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type AddImageResult struct {
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
}

type FileStorage interface {
	GetRestrictedPresignedUploadURL(ctx context.Context, key string, contentType string) (string, error)
}

func (f *Film) Create(
	ctx context.Context,
	name string,
	description string,
) (string, error) {
	const op = "film.Create"

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

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// AddImage validates the content type, verifies that the film exists,
// and returns a presigned upload URL along with the object key in temporary storage.
func (f *Film) AddImage(
	ctx context.Context,
	filmId string,
	contentType string,
) (AddImageResult, error) {
	const op = "film.AddImage"

	log := f.log.With(
		slog.String("op", op),
		slog.String("filmId", filmId),
	)

	log.Info("adding image")

	ext, ok := allowedPictureMIMETypes[contentType]
	if !ok {
		return AddImageResult{}, ErrIncorrectMIMEType
	}

	_, err := f.filmProvider.GetById(ctx, filmId)
	if err != nil {
		log.Error("failed to get film", sl.Err(err))

		return AddImageResult{}, fmt.Errorf("%s: %w", op, ErrFilmNotFound)
	}

	fileName := uuid.New().String() + ext

	key := fmt.Sprintf("tmp/%s/%s", filmId, fileName)

	presignedUrl, err := f.fileStorage.GetRestrictedPresignedUploadURL(ctx, key, contentType)
	if err != nil {
		return AddImageResult{}, err
	}

	return AddImageResult{PresignedURL: presignedUrl, Key: key}, nil
}
