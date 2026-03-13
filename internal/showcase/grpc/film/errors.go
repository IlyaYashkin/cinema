package film

import (
	"cinema/internal/showcase/services/film"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, film.ErrIncorrectMIMEType):
		return status.Error(codes.InvalidArgument, "incorrect mime type")
	case errors.Is(err, film.ErrFilmNotFound):
		return status.Error(codes.NotFound, "film not found")
	}

	return status.Error(codes.Internal, "internal server error")
}
