package auth

import (
	"cinema/gen/sso"
	"net/mail"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateRegisterRequest(in *sso.RegisterRequest) error {
	if err := validateCredentials(in.GetEmail(), in.GetPassword()); err != nil {
		return err
	}
	if len(in.GetPassword()) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	return nil
}

func validateCredentials(email, password string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	return nil
}
