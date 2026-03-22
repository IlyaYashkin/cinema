package media

import (
	grpcMedia "cinema/gen/media"

	"google.golang.org/grpc"
)

type Controller struct {
	grpcMedia.UnimplementedMediaServer
	media Media
}

func NewController(media Media) *Controller {
	return &Controller{
		media: media,
	}
}

func (c *Controller) RegisterGRPCServer(gRPCServer *grpc.Server) {
	grpcMedia.RegisterMediaServer(gRPCServer, c)
}

type Media interface {
}
