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
	Update(
		ctx context.Context,
		filmId string,
		name string,
		description string,
	) error
	Get(
		ctx context.Context,
		id string,
	) (domain.Film, error)
	Delete(
		ctx context.Context,
		filmId string,
	) error
	UploadImage(
		ctx context.Context,
		filmId string,
		contentType string,
	) (film.UploadImageResult, error)
	UpdatePoster(
		ctx context.Context,
		filmId string,
		key string) error
	DeletePoster(
		ctx context.Context,
		filmId string,
	) error
	UpdateImages(
		ctx context.Context,
		filmId string,
		keys []string,
	) (film.UpdateImagesResult, error)
	DeleteImages(
		ctx context.Context,
		filmId string,
		imageIds []int64,
	) (film.DeleteImagesResult, error)
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

func (c *Controller) Update(
	ctx context.Context,
	in *showcase.FilmUpdateRequest,
) (*showcase.FilmUpdateResponse, error) {
	err := c.film.Update(ctx, in.GetFilmId(), in.GetName(), in.GetDescription())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.FilmUpdateResponse{}, nil
}

func (c *Controller) Get(
	ctx context.Context,
	in *showcase.FilmGetRequest,
) (*showcase.FilmGetResponse, error) {
	f, err := c.film.Get(ctx, in.GetFilmId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	images := make([]*showcase.FilmImage, len(f.Images))
	for i, img := range f.Images {
		images[i] = &showcase.FilmImage{
			Id:  img.Id,
			Url: img.Url,
		}
	}

	return &showcase.FilmGetResponse{
		Id:          f.Id.String(),
		Name:        f.Name,
		Description: f.Description,
		PosterUrl:   *f.PosterUrl,
		Images:      images,
	}, nil
}

func (c *Controller) Delete(
	ctx context.Context,
	in *showcase.FilmDeleteRequest,
) (*showcase.FilmDeleteResponse, error) {
	err := c.film.Delete(ctx, in.GetFilmId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.FilmDeleteResponse{}, nil
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

func (c *Controller) DeletePoster(
	ctx context.Context,
	in *showcase.DeletePosterRequest,
) (*showcase.DeletePosterResponse, error) {
	err := c.film.DeletePoster(ctx, in.GetFilmId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.DeletePosterResponse{}, nil
}

func (c *Controller) UpdateImages(
	ctx context.Context,
	in *showcase.UpdateImagesRequest,
) (*showcase.UpdateImagesResponse, error) {
	result, err := c.film.UpdateImages(ctx, in.GetFilmId(), in.GetKeys())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.UpdateImagesResponse{
		FailedKeys:  result.FailedKeys,
		UpdatedKeys: result.UpdatedKeys,
	}, nil
}

func (c *Controller) DeleteImages(
	ctx context.Context,
	in *showcase.DeleteImagesRequest,
) (*showcase.DeleteImagesResponse, error) {
	result, err := c.film.DeleteImages(ctx, in.GetFilmId(), in.GetImageIds())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &showcase.DeleteImagesResponse{
		FailedIds:  result.FailedIds,
		DeletedIds: result.DeletedIds,
	}, nil
}
