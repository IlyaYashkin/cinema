package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")

	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrTokenExpired        = errors.New("token expired")

	ErrPermissionDenied = errors.New("permission denied")
)
