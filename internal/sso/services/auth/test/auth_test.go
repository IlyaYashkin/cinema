package test

import (
	"cinema/internal/sso/domain"
	"cinema/internal/sso/services/auth"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth_Login_InvalidPassword(t *testing.T) {
	email := "test@test.com"
	password := "password123123123_wrong"

	userProvider := NewMockUserProvider(t)
	sessionStorage := NewMockSessionStorage(t)
	tokenGenerator := NewMockTokenGenerator(t)
	resetTokenStorage := NewMockResetTokenStorage(t)
	notificationSender := NewMockNotificationSender(t)

	passHash, _ := bcrypt.GenerateFromPassword([]byte("password123123123"), bcrypt.DefaultCost)
	user := domain.User{
		Id:       uuid.New(),
		Email:    email,
		PassHash: passHash,
	}

	userProvider.On("FindByEmail", mock.Anything, email).Return(user, nil)

	srv := auth.New(
		slog.Default(),
		userProvider,
		sessionStorage,
		resetTokenStorage,
		tokenGenerator,
		notificationSender,
		time.Minute*15,
	)

	_, err := srv.Login(context.Background(), email, password, "", "")

	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}
