package film

import (
	"cinema/gen/showcase"
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
	AddImage(
		ctx context.Context,
		filmId string,
		contentType string,
	) (film.AddImageResult, error)
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
	return &showcase.FilmGetResponse{}, nil
}

func (c *Controller) AddImage(
	ctx context.Context,
	in *showcase.AddImageRequest,
) (*showcase.AddImageResponse, error) {
	res, err := c.film.AddImage(ctx, in.GetFilmId(), in.GetContentType())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.AddImageResponse{PresignedUrl: res.PresignedURL, Key: res.Key}, nil
}
