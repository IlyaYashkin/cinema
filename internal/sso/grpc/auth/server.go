package auth

import (
	"cinema/internal/lib/jwt"
	"cinema/internal/sso/domain"
	"cinema/internal/sso/services/auth"
	"context"
	"errors"
	"net/mail"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ssov1 "cinema/gen/sso"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (tokenPair *domain.TokenPair, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID string, err error)
	Refresh(
		ctx context.Context,
		refreshToken string,
	) (tokenPair *jwt.TokenPair, err error)
	Logout(
		ctx context.Context,
		refreshToken string,
	) (err error)
}

func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if err := validateLoginRequest(in); err != nil {
		return nil, err
	}

	tokenPair, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov1.LoginResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegisterRequest(in); err != nil {
		return nil, err
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) Refresh(
	ctx context.Context,
	in *ssov1.RefreshRequest,
) (*ssov1.RefreshResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := s.auth.Refresh(ctx, in.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}

		return nil, status.Error(codes.Internal, "failed to refresh")
	}

	return &ssov1.RefreshResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

func (s *serverAPI) Logout(
	ctx context.Context,
	in *ssov1.LogoutRequest,
) (*ssov1.LogoutResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	err := s.auth.Logout(ctx, in.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}

		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssov1.LogoutResponse{}, nil
}

func validateLoginRequest(in *ssov1.LoginRequest) error {
	if in.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	return nil
}

func validateRegisterRequest(in *ssov1.RegisterRequest) error {
	if in.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	if len(in.Password) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	return nil
}
