package auth

import (
	"cinema/internal/sso/services/auth"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid email or password")
	case errors.Is(err, auth.ErrInvalidPassword):
		return status.Error(codes.Unauthenticated, "invalid password")
	case errors.Is(err, auth.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, auth.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, auth.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, "token expired")
	case errors.Is(err, auth.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, "invalid refresh token")
	case errors.Is(err, auth.ErrInvalidAccessToken):
		return status.Error(codes.Unauthenticated, "invalid access token")

	case errors.Is(err, auth.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	}

	return status.Error(codes.Internal, "internal server error")
}
