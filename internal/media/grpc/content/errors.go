package content

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	return status.Error(codes.Internal, "internal server error")
}
