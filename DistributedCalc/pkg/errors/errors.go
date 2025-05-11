package errors

import (
	"encoding/json"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewBadRequestError(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func NewNotFoundError(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg}
}

func NewInternalError(msg string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg}
}

func HandleHTTPError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if appErr, ok := err.(*AppError); ok {
		w.WriteHeader(appErr.Code)
		json.NewEncoder(w).Encode(appErr)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(&AppError{Code: http.StatusInternalServerError, Message: "internal server error"})
}
