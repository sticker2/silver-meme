package tasks

import (
	"DistributedCalc/pkg/errors"
	"net/http"
)

func NewTaskFetchError(msg string) *errors.AppError {
	return &errors.AppError{Code: http.StatusInternalServerError, Message: msg}
}

func NewTaskSubmitError() *errors.AppError {
	return &errors.AppError{Code: http.StatusInternalServerError, Message: "failed to submit task result"}
}

func NewInvalidTaskError(msg string) *errors.AppError {
	return &errors.AppError{Code: http.StatusBadRequest, Message: msg}
}

func NewTaskNotFoundError() *errors.AppError {
	return &errors.AppError{Code: http.StatusNotFound, Message: "no pending tasks"}
}
