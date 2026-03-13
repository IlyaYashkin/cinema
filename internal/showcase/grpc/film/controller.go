package film

import (
	"cinema/gen/showcase"
	"cinema/internal/showcase/domain"
	"cinema/internal/showcase/services/film"
	"context"

	"google.golang.org/grpc"
)

type Controller struct {
	showcase.UnimplementedFilmServer
	film Film
}

func NewController(film Film) *Controller {
	return &Controller{
		film: film,
	}
}

func (c *Controller) RegisterGRPCServer(gRPCServer *grpc.Server) {
	showcase.RegisterFilmServer(gRPCServer, c)
}

type Film interface {
	Create(
		ctx context.Context,
		name string,
		description string,
	) (string, error)
	Get(
		ctx context.Context,
		id string,
	) (domain.Film, error)
	UploadImage(
		ctx context.Context,
		filmId string,
		contentType string,
	) (film.UploadImageResult, error)
	UpdatePoster(
		ctx context.Context,
		filmId string,
		key string) error
}

func (c *Controller) Create(
	ctx context.Context,
	in *showcase.FilmCreateRequest,
) (*showcase.FilmCreateResponse, error) {
	id, err := c.film.Create(ctx, in.GetName(), in.GetDescription())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.FilmCreateResponse{Id: id}, nil
}

func (c *Controller) Get(
	ctx context.Context,
	in *showcase.FilmGetRequest,
) (*showcase.FilmGetResponse, error) {
	f, err := c.film.Get(ctx, in.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.FilmGetResponse{
		Id:          f.Id.String(),
		Name:        f.Name,
		Description: f.Description,
		PosterUrl:   *f.PosterUrl,
	}, nil
}

func (c *Controller) UploadImage(
	ctx context.Context,
	in *showcase.UploadImageRequest,
) (*showcase.UploadImageResponse, error) {
	res, err := c.film.UploadImage(ctx, in.GetFilmId(), in.GetContentType())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.UploadImageResponse{PresignedUrl: res.PresignedURL, Key: res.Key}, nil
}

func (c *Controller) UpdatePoster(
	ctx context.Context,
	in *showcase.UpdatePosterRequest,
) (*showcase.UpdatePosterResponse, error) {
	err := c.film.UpdatePoster(ctx, in.GetFilmId(), in.GetKey())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.UpdatePosterResponse{}, nil
}
