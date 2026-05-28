package content

import (
	"cinema/gen/media"
	"cinema/internal/media/service/content"
	"context"

	"google.golang.org/grpc"
)

type Controller struct {
	media.UnimplementedContentServer
	srv Service
}

func NewController(media Service) *Controller {
	return &Controller{
		srv: media,
	}
}

func (c *Controller) RegisterGRPCServer(gRPCServer *grpc.Server) {
	media.RegisterContentServer(gRPCServer, c)
}

type Service interface {
	InitUpload(
		ctx context.Context,
		filmId string,
		fileName string,
		contentType string,
		fileSize int64,
	) (content.InitVideoUploadResult, error)
	CompleteUpload(
		ctx context.Context,
		uploadId string,
		key string,
		eTags []string,
	) error
	AbortUpload(
		ctx context.Context,
		uploadId string,
		key string,
	) error
}

func (c *Controller) InitUpload(
	ctx context.Context,
	in *media.InitUploadRequest,
) (*media.InitUploadResponse, error) {
	result, err := c.srv.InitUpload(ctx, in.GetFilmId(), in.GetFileName(), in.GetContentType(), in.GetFileSize())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &media.InitUploadResponse{
		UploadId:      result.UploadId,
		Key:           result.Key,
		PresignedUrls: result.PresignedURLs,
		ChunkSize:     result.ChunkSize,
	}, nil
}

func (c *Controller) CompleteUpload(
	ctx context.Context,
	in *media.CompleteUploadRequest,
) (*media.CompleteUploadResponse, error) {
	err := c.srv.CompleteUpload(ctx, in.GetUploadId(), in.GetKey(), in.GetETags())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &media.CompleteUploadResponse{}, nil
}

func (c *Controller) AbortUpload(
	ctx context.Context,
	in *media.AbortUploadRequest,
) (*media.AbortUploadResponse, error) {
	err := c.srv.AbortUpload(ctx, in.GetUploadId(), in.GetKey())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &media.AbortUploadResponse{}, nil
}
