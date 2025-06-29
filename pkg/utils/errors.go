// Description: Custom error handling for validation errors in Go applications.
package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type ErrorType string

const (
	ErrValidation          ErrorType = "validation_error"
	ErrRefreshTokenExpired ErrorType = "refresh_token_expired"
	ErrInvalidRefreshToken ErrorType = "invalid_refresh_token"
)

type AppError struct {
	Type    ErrorType
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewRefreshTokenExpired() *AppError {
	return &AppError{
		Type:    ErrRefreshTokenExpired,
		Message: "refresh token expired",
	}
}

func NewInvalidRefreshToken() *AppError {
	return &AppError{
		Type:    ErrInvalidRefreshToken,
		Message: "invalid refresh token",
	}
}

func CustomValidationErrors(errs validator.ValidationErrors) []string {
	var messages []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", err.Field()))
		case "email":
			messages = append(messages, fmt.Sprintf("%s must be a valid email address", err.Field()))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s is not valid", err.Field()))
		}
	}
	return messages
}