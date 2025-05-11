package storage

import (
	"DistributedCalc/pkg/errors"
	"net/http"
)

func NewDBError(msg string) *errors.AppError {
	return &errors.AppError{Code: http.StatusInternalServerError, Message: msg}
}

func NewUserNotFoundError() *errors.AppError {
	return &errors.AppError{Code: http.StatusNotFound, Message: "user not found"}
}

func NewUserExistsError() *errors.AppError {
	return &errors.AppError{Code: http.StatusBadRequest, Message: "user already exists"}
}

func NewTaskNotFoundError() *errors.AppError {
	return &errors.AppError{Code: http.StatusNotFound, Message: "no pending tasks"}
}
