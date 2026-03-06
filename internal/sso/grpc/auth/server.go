package auth

import (
	"cinema/internal/sso/domain"
	"cinema/internal/sso/services/auth"
	"context"
	"errors"
	"net/mail"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cinema/gen/sso"
)

type Controller struct {
	sso.UnimplementedAuthServer
	auth Auth
}

func NewController(auth Auth) *Controller {
	return &Controller{
		auth: auth,
	}
}

func (c *Controller) RegisterGRPCServer(gRPCServer *grpc.Server) {
	sso.RegisterAuthServer(gRPCServer, c)
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
	) (tokenPair *domain.TokenPair, err error)
	Logout(
		ctx context.Context,
		refreshToken string,
	) (err error)
}

func (c *Controller) Login(
	ctx context.Context,
	in *sso.LoginRequest,
) (*sso.LoginResponse, error) {
	if err := validateLoginRequest(in); err != nil {
		return nil, err
	}

	tokenPair, err := c.auth.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &sso.LoginResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

func (c *Controller) Register(
	ctx context.Context,
	in *sso.RegisterRequest,
) (*sso.RegisterResponse, error) {
	if err := validateRegisterRequest(in); err != nil {
		return nil, err
	}

	uid, err := c.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &sso.RegisterResponse{UserId: uid}, nil
}

func (c *Controller) Refresh(
	ctx context.Context,
	in *sso.RefreshRequest,
) (*sso.RefreshResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := c.auth.Refresh(ctx, in.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}

		return nil, status.Error(codes.Internal, "failed to refresh")
	}

	return &sso.RefreshResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

func (c *Controller) Logout(
	ctx context.Context,
	in *sso.LogoutRequest,
) (*sso.LogoutResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	err := c.auth.Logout(ctx, in.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}

		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &sso.LogoutResponse{}, nil
}

func validateLoginRequest(in *sso.LoginRequest) error {
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

func validateRegisterRequest(in *sso.RegisterRequest) error {
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
