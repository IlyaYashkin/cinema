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

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type SessionStorage interface {
	Save(ctx context.Context, userId, token string, ttl time.Duration) error
	Exists(ctx context.Context, userId, token string) (bool, error)
	Delete(ctx context.Context, userId, token string) error
	DeleteAll(ctx context.Context, userId string) error
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (id string, err error)
}

type UserProvider interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type TokenGenerator interface {
	GenerateAccessToken(userId string) (string, error)
	GenerateRefreshToken(userId string) (string, error)
	ValidateToken(tokenString string) (*jwt.Claims, error)
	GetRefreshTTL() time.Duration
}

type Auth struct {
	log            *slog.Logger
	userSaver      UserSaver
	usrProvider    UserProvider
	sessionStorage SessionStorage
	jwtGenerator   TokenGenerator
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	sessionStorage SessionStorage,
	generator TokenGenerator,
) *Auth {
	return &Auth{
		log:            log,
		userSaver:      userSaver,
		usrProvider:    userProvider,
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

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
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

	user, err := a.usrProvider.FindByEmail(ctx, email)
	if err != nil {
		log.Error("failed to login user", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
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

	accessToken, err := a.jwtGenerator.GenerateAccessToken(userId)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := a.jwtGenerator.GenerateRefreshToken(userId)
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

	accessToken, err := a.jwtGenerator.GenerateAccessToken(userId)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err = a.jwtGenerator.GenerateRefreshToken(userId)
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

	log.Info("user logged out")

	return nil
}
