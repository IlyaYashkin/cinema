package auth

import (
	"cinema/internal/sso/domain"
	"cinema/internal/sso/services/auth"
	"context"
	"errors"
	"net/mail"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	) (userId string, err error)
	Refresh(
		ctx context.Context,
		refreshToken string,
	) (tokenPair *domain.TokenPair, err error)
	Logout(
		ctx context.Context,
		refreshToken string,
	) (err error)
	ChangeRole(
		ctx context.Context,
		token string,
		userId string,
		role domain.Role,
	) (err error)
}

func (c *Controller) Login(
	ctx context.Context,
	in *sso.LoginRequest,
) (*sso.LoginResponse, error) {
	if err := validateCredentials(in.GetEmail(), in.GetPassword()); err != nil {
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
	if in.GetRefreshToken() == "" {
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
	if in.GetRefreshToken() == "" {
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

func (c *Controller) ChangeRole(
	ctx context.Context,
	in *sso.ChangeRoleRequest,
) (*sso.ChangeRoleResponse, error) {
	token, err := getBearerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if in.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	role := domain.Role(in.GetRole())

	if !role.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	err = c.auth.ChangeRole(ctx, token, in.GetUserId(), role)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidAccessToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid access token")
		}
		if errors.Is(err, auth.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to set role")
	}

	return &sso.ChangeRoleResponse{}, nil
}

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

func validateRegisterRequest(in *sso.RegisterRequest) error {
	if err := validateCredentials(in.GetEmail(), in.GetPassword()); err != nil {
		return err
	}
	if len(in.GetPassword()) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	return nil
}

func validateCredentials(email, password string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	return nil
}
