package jwt

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/sl"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken            = errors.New("invalid token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type Generator struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewGenerator(config config.JWTConfig) (*Generator, error) {
	const op = "lib.jwt.new_generator"

	privateKeyPEM, err := os.ReadFile(config.PrivateKeyPath)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	publicKeyPEM, err := os.ReadFile(config.PublicKeyPath)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	return &Generator{
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  config.AccessTTL,
		refreshTTL: config.RefreshTTL,
	}, nil
}

func (g *Generator) GenerateAccessToken(userId string, role string) (string, error) {
	const op = "lib.jwt.generate_access_token"

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.accessTTL)),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", sl.WrapErr(op, err)
	}

	return signed, nil
}

func (g *Generator) GenerateRefreshToken(userId string, role string) (string, error) {
	const op = "lib.jwt.generate_refresh_token"

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.refreshTTL)),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", sl.WrapErr(op, err)
	}

	return signed, nil
}

func (g *Generator) ValidateToken(tokenString string) (*Claims, error) {
	const op = "lib.jwt.validate_token"

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, sl.WrapErr(op, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"]))
		}
		return g.publicKey, nil
	})
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, sl.WrapErr(op, ErrInvalidToken)
	}

	return claims, nil
}

func (g *Generator) GetRefreshTTL() time.Duration {
	return g.refreshTTL
}
