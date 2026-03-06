package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrUserNotFound        = errors.New("user not found")
	ErrPermissionDenied    = errors.New("permission denied")
)
