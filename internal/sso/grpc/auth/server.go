package auth

import (
	"cinema/internal/sso/domain"
	"context"

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
		accessToken string,
		userId string,
		role domain.Role,
	) (err error)
	ChangeEmail(
		ctx context.Context,
		accessToken string,
		newEmail string,
		password string,
	) (err error)
	ChangePassword(
		ctx context.Context,
		accessToken string,
		oldPassword string,
		newPassword string,
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
		return nil, toGRPCError(err)
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
		return nil, toGRPCError(err)
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
		return nil, toGRPCError(err)
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
		return nil, toGRPCError(err)
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
		return nil, toGRPCError(err)
	}

	return &sso.ChangeRoleResponse{}, nil
}

func (c *Controller) ChangeEmail(
	ctx context.Context,
	in *sso.ChangeEmailRequest,
) (*sso.ChangeEmailResponse, error) {
	token, err := getBearerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetNewEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "new email is required")
	}
	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	err = c.auth.ChangeEmail(ctx, token, in.GetNewEmail(), in.GetPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &sso.ChangeEmailResponse{}, nil
}

func (c *Controller) ChangePassword(
	ctx context.Context,
	in *sso.ChangePasswordRequest,
) (*sso.ChangePasswordResponse, error) {
	token, err := getBearerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetOldPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "old password is required")
	}
	if in.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "new password is required")
	}

	err = c.auth.ChangePassword(ctx, token, in.GetOldPassword(), in.GetNewPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &sso.ChangePasswordResponse{}, nil
}
