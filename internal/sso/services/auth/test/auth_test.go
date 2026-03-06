package test

import (
	"cinema/internal/sso/domain"
	auth2 "cinema/internal/sso/services/auth"
	"context"
	"log/slog"
	"testing"

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

	passHash, _ := bcrypt.GenerateFromPassword([]byte("password123123123"), bcrypt.DefaultCost)
	user := domain.User{
		Id:       uuid.New(),
		Email:    email,
		PassHash: passHash,
	}

	userProvider.On("FindByEmail", mock.Anything, email).Return(user, nil)

	auth := auth2.New(slog.Default(), nil, userProvider, sessionStorage, tokenGenerator)

	_, err := auth.Login(context.Background(), email, password)

	require.ErrorIs(t, err, auth2.ErrInvalidCredentials)
}
