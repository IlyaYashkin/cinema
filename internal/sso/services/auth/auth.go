package auth

import (
	"cinema/internal/lib/jwt"
	"cinema/internal/lib/sl"
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SessionStorage interface {
	Save(ctx context.Context, userId, token string, ttl time.Duration) error
	Exists(ctx context.Context, userId, token string) (bool, error)
	Delete(ctx context.Context, userId, token string) error
	DeleteAll(ctx context.Context, userId string) error
}

type UserProvider interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (id string, err error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	ChangeUserRole(ctx context.Context, userId string, role string) error
}

type TokenGenerator interface {
	GenerateAccessToken(userId string, role string) (string, error)
	GenerateRefreshToken(userId string, role string) (string, error)
	ValidateToken(tokenString string) (*jwt.Claims, error)
	GetRefreshTTL() time.Duration
}

type Auth struct {
	log            *slog.Logger
	userProvider   UserProvider
	sessionStorage SessionStorage
	jwtGenerator   TokenGenerator
}

func New(
	log *slog.Logger,
	userProvider UserProvider,
	sessionStorage SessionStorage,
	generator TokenGenerator,
) *Auth {
	return &Auth{
		log:            log,
		userProvider:   userProvider,
		sessionStorage: sessionStorage,
		jwtGenerator:   generator,
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
		log.Error("failed to save user", sl.Err(err))

		if errors.Is(err, storage.ErrUserExists) {
			return "", ErrUserAlreadyExists
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Login(ctx context.Context, email string, password string) (*domain.TokenPair, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("logging user")

	user, err := a.userProvider.FindByEmail(ctx, email)
	if err != nil {
		log.Error("failed to login user", sl.Err(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error("invalid password", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	userId := user.Id.String()

	err = a.sessionStorage.DeleteAll(ctx, userId)
	if err != nil {
		log.Error("failed to delete old sessions", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

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

	err = a.sessionStorage.Save(ctx, userId, refreshToken, a.jwtGenerator.GetRefreshTTL())
	if err != nil {
		log.Error("failed to save refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	const op = "Auth.Refresh"

	log := a.log.With(
		slog.String("op", op),
	)

	claims, err := a.jwtGenerator.ValidateToken(refreshToken)
	if err != nil {
		log.Error("failed to validate refresh token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	userId := claims.Subject
	role := claims.Role

	log = log.With(
		slog.String("userId", userId),
	)

	exists, err := a.sessionStorage.Exists(ctx, userId, refreshToken)
	if err != nil {
		log.Error("failed to check session", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		log.Warn("refresh token not found")

		return nil, fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	if err := a.sessionStorage.Delete(ctx, userId, refreshToken); err != nil {
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

	if err := a.sessionStorage.Save(ctx, userId, refreshToken, a.jwtGenerator.GetRefreshTTL()); err != nil {
		log.Error("failed to save refresh token", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	const op = "Auth.Logout"

	log := a.log.With(
		slog.String("op", op),
	)

	claims, err := a.jwtGenerator.ValidateToken(refreshToken)
	if err != nil {
		log.Error("failed to validate refresh token", sl.Err(err))

		return fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	userId := claims.Subject

	log = log.With(
		slog.String("userId", userId),
	)

	exists, err := a.sessionStorage.Exists(ctx, userId, refreshToken)
	if err != nil {
		log.Error("failed to check session", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		log.Warn("refresh token not found")

		return fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	err = a.sessionStorage.DeleteAll(ctx, userId)
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

	claims, err := a.jwtGenerator.ValidateToken(accessToken)
	if err != nil {
		log.Error("failed to validate access token", sl.Err(err))

		return fmt.Errorf("%s: %w", op, ErrInvalidAccessToken)
	}

	log = log.With(
		slog.String("changed by", claims.Subject),
		slog.String("changed by (role)", claims.Role),
	)

	if domain.Role(claims.Role) != domain.Admin {
		log.Error("attempt to change role without permission")

		return fmt.Errorf("%s: %w", op, ErrPermissionDenied)
	}

	err = a.userProvider.ChangeUserRole(ctx, userId, string(role))
	if err != nil {
		log.Error("failed to set user role", sl.Err(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("invalidating user session after role change")

	err = a.sessionStorage.DeleteAll(ctx, userId)
	if err != nil {
		log.Error("failed to delete sessions", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user session invalidated")

	log.Info("user role changed")

	return nil
}
