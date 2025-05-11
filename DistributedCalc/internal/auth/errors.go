package auth

import (
	"DistributedCalc/pkg/errors"
	"net/http"
)

func NewBadRequestError(msg string) *errors.AppError {
	return &errors.AppError{Code: http.StatusBadRequest, Message: msg}
}

func NewInternalError(msg string) *errors.AppError {
	return &errors.AppError{Code: http.StatusInternalServerError, Message: msg}
}

func NewUserExistsError() *errors.AppError {
	return &errors.AppError{Code: http.StatusBadRequest, Message: "user already exists"}
}

func NewInvalidCredentialsError() *errors.AppError {
	return &errors.AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
}

func NewTokenGenerationError() *errors.AppError {
	return &errors.AppError{Code: http.StatusInternalServerError, Message: "failed to generate token"}
}

func NewInvalidTokenError() *errors.AppError {
	return &errors.AppError{Code: http.StatusUnauthorized, Message: "invalid token"}
}
