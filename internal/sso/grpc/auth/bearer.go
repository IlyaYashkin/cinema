package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func getBearerFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header format")
	}
	token := authHeader[7:]

	if token == "" {
		return "", status.Error(codes.Unauthenticated, "missing token")
	}

	return token, nil
}
