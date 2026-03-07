package auth

import (
	"cinema/internal/sso/domain"
	"net/mail"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateEmail(email, fieldName string) error {
	if email == "" {
		return status.Errorf(codes.InvalidArgument, "%s is required", fieldName)
	}
	if err := validateMaxLength(email, fieldName, 254); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	return nil
}

func validatePassword(password, fieldName string) error {
	if password == "" {
		return status.Errorf(codes.InvalidArgument, "%s is required", fieldName)
	}
	if err := validateMinLength(password, fieldName, 8); err != nil {
		return err
	}
	if err := validateMaxLength(password, fieldName, 72); err != nil {
		return err
	}
	return nil
}

func validateDeviceId(deviceId, fieldName string) error {
	if deviceId == "" {
		return status.Errorf(codes.InvalidArgument, "%s is required", fieldName)
	}
	if err := validateMaxLength(deviceId, fieldName, 64); err != nil {
		return err
	}
	return nil
}

func validateRole(role string) error {
	if role == "" {
		return status.Error(codes.InvalidArgument, "role is required")
	}

	domainRole := domain.Role(role)

	if !domainRole.IsValid() {
		return status.Error(codes.InvalidArgument, "invalid role")
	}
	return nil
}

func validateResetToken(resetToken, fieldName string) error {
	if resetToken == "" {
		return status.Errorf(codes.InvalidArgument, "%s is required", fieldName)
	}
	if err := validateMaxLength(resetToken, fieldName, 64); err != nil {
		return err
	}
	return nil
}

func validateMinLength(value, field string, min int) error {
	if len(value) < min {
		return status.Errorf(codes.InvalidArgument, "%s must be at least %d characters", field, min)
	}
	return nil
}

func validateMaxLength(value, field string, max int) error {
	if len(value) > max {
		return status.Errorf(codes.InvalidArgument, "%s must be at most %d characters", field, max)
	}
	return nil
}
