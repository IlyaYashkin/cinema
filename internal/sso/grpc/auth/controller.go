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
	Login(ctx context.Context, email string, password string, deviceId string, deviceName string) (tokenPair *domain.TokenPair, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userId string, err error)
	Refresh(ctx context.Context, refreshToken string, deviceId string, deviceName string) (tokenPair *domain.TokenPair, err error)
	Logout(ctx context.Context, refreshToken string, deviceId string) (err error)
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
	ForgotPassword(
		ctx context.Context,
		email string,
	)
	ResetPassword(
		ctx context.Context,
		resetToken string,
		newPassword string,
	) (err error)
}

func (c *Controller) Login(
	ctx context.Context,
	in *sso.LoginRequest,
) (*sso.LoginResponse, error) {
	if err := validateEmail(in.GetEmail(), "email"); err != nil {
		return nil, err
	}
	if err := validatePassword(in.GetPassword(), "password"); err != nil {
		return nil, err
	}
	if err := validateDeviceId(in.GetDeviceId(), "device_id"); err != nil {
		return nil, err
	}

	userAgent := getUserAgent(ctx)

	tokenPair, err := c.auth.Login(ctx, in.GetEmail(), in.GetPassword(), in.GetDeviceId(), userAgent)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &sso.LoginResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

func (c *Controller) Register(
	ctx context.Context,
	in *sso.RegisterRequest,
) (*sso.RegisterResponse, error) {
	if err := validateEmail(in.GetEmail(), "email"); err != nil {
		return nil, err
	}
	if err := validatePassword(in.GetPassword(), "password"); err != nil {
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
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}
	if err := validateDeviceId(in.GetDeviceId(), "device_id"); err != nil {
		return nil, err
	}

	userAgent := getUserAgent(ctx)

	tokenPair, err := c.auth.Refresh(ctx, in.GetRefreshToken(), in.GetDeviceId(), userAgent)
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
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}
	if err := validateDeviceId(in.GetDeviceId(), "device_id"); err != nil {
		return nil, err
	}

	err := c.auth.Logout(ctx, in.GetRefreshToken(), in.GetDeviceId())
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
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if err := validateRole(in.GetRole()); err != nil {
		return nil, err
	}

	err = c.auth.ChangeRole(ctx, token, in.GetUserId(), domain.Role(in.GetRole()))
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

	if err := validateEmail(in.GetNewEmail(), "new_email"); err != nil {
		return nil, err
	}
	if err := validatePassword(in.GetPassword(), "password"); err != nil {
		return nil, err
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

	if err := validatePassword(in.GetOldPassword(), "old_password"); err != nil {
		return nil, err
	}
	if err := validatePassword(in.GetNewPassword(), "new_password"); err != nil {
		return nil, err
	}

	err = c.auth.ChangePassword(ctx, token, in.GetOldPassword(), in.GetNewPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &sso.ChangePasswordResponse{}, nil
}

func (c *Controller) ForgotPassword(
	ctx context.Context,
	in *sso.ForgotPasswordRequest,
) (*sso.ForgotPasswordResponse, error) {
	if err := validateEmail(in.GetEmail(), "email"); err != nil {
		return nil, err
	}

	c.auth.ForgotPassword(ctx, in.GetEmail())

	return &sso.ForgotPasswordResponse{}, nil
}

func (c *Controller) ResetPassword(
	ctx context.Context,
	in *sso.ResetPasswordRequest,
) (*sso.ResetPasswordResponse, error) {
	if err := validateResetToken(in.GetNewPassword(), "reset_token"); err != nil {
		return nil, err
	}
	if err := validatePassword(in.GetNewPassword(), "new_password"); err != nil {
		return nil, err
	}

	err := c.auth.ResetPassword(ctx, in.GetResetToken(), in.GetNewPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &sso.ResetPasswordResponse{}, nil
}
