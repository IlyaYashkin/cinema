package auth

import (
	"cinema/internal/lib/jwt"
	"cinema/internal/lib/sl"
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type SessionStorage interface {
	SaveSession(ctx context.Context, userId, token, deviceId, deviceName string, ttl time.Duration) error
	GetSession(ctx context.Context, userId, deviceId string) (domain.Session, error)
	DeleteSession(ctx context.Context, userId, deviceId string) error
	DeleteAllSessions(ctx context.Context, userId string) error
}

type ResetTokenStorage interface {
	SaveResetToken(ctx context.Context, userId, token string, ttl time.Duration) error
	GetUserIdByResetToken(ctx context.Context, token string) (string, error)
	DeleteResetToken(ctx context.Context, token string) error
}

type UserProvider interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (id string, err error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id string) (domain.User, error)
	UpdateUserRole(ctx context.Context, userId string, role string) error
	UpdateUserEmail(ctx context.Context, userId string, email string) error
	UpdateUserPassword(ctx context.Context, userId string, passHash []byte) error
}

type TokenGenerator interface {
	GenerateAccessToken(userId string, role string) (string, error)
	GenerateRefreshToken(userId string, role string) (string, error)
	ValidateToken(tokenString string) (*jwt.Claims, error)
	GetRefreshTTL() time.Duration
}

type NotificationSender interface {
	SendPasswordResetNotification(ctx context.Context, email, resetToken string) error
}

type Auth struct {
	log                *slog.Logger
	userProvider       UserProvider
	sessionStorage     SessionStorage
	resetTokenStorage  ResetTokenStorage
	jwtGenerator       TokenGenerator
	notificationSender NotificationSender
	resetTokenTTL      time.Duration
}

func New(
	log *slog.Logger,
	userProvider UserProvider,
	sessionStorage SessionStorage,
	resetTokenStorage ResetTokenStorage,
	generator TokenGenerator,
	notificationSender NotificationSender,
	resetTokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:                log,
		userProvider:       userProvider,
		sessionStorage:     sessionStorage,
		resetTokenStorage:  resetTokenStorage,
		jwtGenerator:       generator,
		notificationSender: notificationSender,
		resetTokenTTL:      resetTokenTTL,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (string, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userProvider.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return "", ErrUserAlreadyExists
		}

		log.Error("failed to save user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Login(ctx context.Context, email string, password string, deviceId string, deviceName string) (*domain.TokenPair, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.String("deviceId", deviceId),
		slog.String("deviceName", deviceName),
	)

	log.Info("logging user")

	user, err := a.userProvider.FindByEmail(ctx, email)
	if err != nil {
		log.Error("failed to find user", sl.Err(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid password", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	userId := user.Id.String()

	accessToken, err := a.jwtGenerator.GenerateAccessToken(user.Id.String(), string(user.Role))
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := a.jwtGenerator.GenerateRefreshToken(userId, string(user.Role))
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.SaveSession(ctx, userId, refreshToken, deviceId, deviceName, a.jwtGenerator.GetRefreshTTL())
	if err != nil {
		log.Error("failed to save refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string, deviceId string, deviceName string) (*domain.TokenPair, error) {
	const op = "Auth.Refresh"

	log := a.log.With(
		slog.String("op", op),
		slog.String("deviceId", deviceId),
		slog.String("deviceName", deviceName),
	)

	claims, err := a.validateRefreshToken(log, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userId := claims.Subject
	role := claims.Role

	log = log.With(
		slog.String("userId", userId),
	)

	session, err := a.sessionStorage.GetSession(ctx, userId, deviceId)
	if err != nil {
		log.Error("failed to get session", sl.Err(err))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if session.RefreshToken != refreshToken {
		log.Warn("refresh token reuse detected, invalidating all sessions")

		_ = a.sessionStorage.DeleteAllSessions(ctx, userId)

		return nil, fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	if err := a.sessionStorage.DeleteSession(ctx, userId, deviceId); err != nil {
		log.Error("failed to delete old refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := a.jwtGenerator.GenerateAccessToken(userId, role)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err = a.jwtGenerator.GenerateRefreshToken(userId, role)
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := a.sessionStorage.SaveSession(ctx, userId, refreshToken, deviceId, deviceName, a.jwtGenerator.GetRefreshTTL()); err != nil {
		log.Error("failed to save refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *Auth) Logout(ctx context.Context, refreshToken string, deviceId string) error {
	const op = "Auth.Logout"

	log := a.log.With(
		slog.String("op", op),
	)

	claims, err := a.validateRefreshToken(log, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userId := claims.Subject

	log = log.With(
		slog.String("userId", userId),
	)

	session, err := a.sessionStorage.GetSession(ctx, userId, deviceId)
	if err != nil {
		log.Error("failed to get session", sl.Err(err))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if session.RefreshToken != refreshToken {
		log.Warn("refresh token reuse detected, invalidating all sessions")

		_ = a.sessionStorage.DeleteAllSessions(ctx, userId)

		return fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	err = a.sessionStorage.DeleteSession(ctx, userId, deviceId)
	if err != nil {
		log.Error("failed to delete sessions", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) ChangeRole(ctx context.Context, accessToken string, userId string, role domain.Role) error {
	const op = "Auth.ChangeRole"

	log := a.log.With(
		slog.String("op", op),
		slog.String("userId", userId),
		slog.String("role", string(role)),
	)

	log.Info("change user role")

	claims, err := a.validateAccessToken(log, accessToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log = log.With(
		slog.String("changed by", claims.Subject),
		slog.String("changed by (role)", claims.Role),
	)

	if domain.Role(claims.Role) != domain.Admin {
		log.Error("attempt to change role without permission")

		return fmt.Errorf("%s: %w", op, ErrPermissionDenied)
	}

	err = a.userProvider.UpdateUserRole(ctx, userId, string(role))
	if err != nil {
		log.Error("failed to set user role", sl.Err(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.DeleteAllSessions(ctx, userId)
	if err != nil {
		log.Error("failed to delete sessions", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) ChangeEmail(
	ctx context.Context,
	accessToken string,
	newEmail string,
	password string,
) error {
	const op = "Auth.ChangeEmail"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("change user email")

	claims, err := a.validateAccessToken(log, accessToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userId := claims.Subject

	log = log.With(
		slog.String("userId", userId),
	)

	user, err := a.userProvider.FindById(ctx, userId)
	if err != nil {
		log.Error("failed to find user", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid password", sl.Err(err))

		return fmt.Errorf("%s: %w", op, ErrInvalidPassword)
	}

	err = a.userProvider.UpdateUserEmail(ctx, userId, newEmail)
	if err != nil {
		log.Error("failed to change user email", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.DeleteAllSessions(ctx, userId)
	if err != nil {
		log.Error("failed to delete sessions", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) ChangePassword(
	ctx context.Context,
	accessToken string,
	oldPassword string,
	newPassword string,
) error {
	const op = "Auth.ChangePassword"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("change user password")

	claims, err := a.validateAccessToken(log, accessToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userId := claims.Subject

	log = log.With(
		slog.String("userId", userId),
	)

	user, err := a.userProvider.FindById(ctx, userId)
	if err != nil {
		log.Error("failed to find user", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(oldPassword)); err != nil {
		log.Warn("invalid password", sl.Err(err))

		return fmt.Errorf("%s: %w", op, ErrInvalidPassword)
	}

	err = a.generatePasswordHashAndUpdateUser(ctx, log, userId, newPassword)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.DeleteAllSessions(ctx, userId)
	if err != nil {
		log.Error("failed to delete all sessions", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) ForgotPassword(
	ctx context.Context,
	email string,
) {
	const op = "Auth.ForgotPassword"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("forgot user password")

	user, err := a.userProvider.FindByEmail(ctx, email)
	if err != nil {
		log.Error("failed to find user", sl.Err(err))

		return
	}

	log = log.With(
		slog.String("userId", user.Id.String()),
	)

	resetToken := generateResetToken()

	err = a.resetTokenStorage.SaveResetToken(ctx, user.Id.String(), resetToken, a.resetTokenTTL)
	if err != nil {
		log.Error("failed to save reset token", sl.Err(err))

		return
	}

	err = a.notificationSender.SendPasswordResetNotification(ctx, email, resetToken)
	if err != nil {
		log.Error("failed to send password reset notification", sl.Err(err))

		_ = a.resetTokenStorage.DeleteResetToken(ctx, resetToken)

		return
	}
}

func (a *Auth) ResetPassword(
	ctx context.Context,
	resetToken string,
	newPassword string,
) error {
	const op = "Auth.ResetPassword"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("reset user password")

	userId, err := a.resetTokenStorage.GetUserIdByResetToken(ctx, resetToken)
	if err != nil {
		log.Error("failed to find user by reset token", sl.Err(err))

		return ErrInvalidResetToken
	}

	err = a.generatePasswordHashAndUpdateUser(ctx, log, userId, newPassword)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.resetTokenStorage.DeleteResetToken(ctx, resetToken)
	if err != nil {
		log.Error("failed to delete reset token", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.DeleteAllSessions(ctx, userId)
	if err != nil {
		log.Error("failed to delete all sessions", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) validateAccessToken(log *slog.Logger, token string) (*jwt.Claims, error) {
	return a.validateToken(log, token, ErrInvalidAccessToken)
}

func (a *Auth) validateRefreshToken(log *slog.Logger, token string) (*jwt.Claims, error) {
	return a.validateToken(log, token, ErrInvalidRefreshToken)
}

func (a *Auth) validateToken(log *slog.Logger, token string, invalidErr error) (*jwt.Claims, error) {
	claims, err := a.jwtGenerator.ValidateToken(token)
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			log.Warn("token expired")
			return nil, ErrTokenExpired
		}
		log.Error("failed to validate token", sl.Err(err))
		return nil, invalidErr
	}
	return claims, nil
}

func (a *Auth) generatePasswordHashAndUpdateUser(ctx context.Context, log *slog.Logger, userId, password string) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return err
	}

	err = a.userProvider.UpdateUserPassword(ctx, userId, passHash)
	if err != nil {
		log.Error("failed to update user password", sl.Err(err))

		return err
	}

	return nil
}

func generateResetToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
