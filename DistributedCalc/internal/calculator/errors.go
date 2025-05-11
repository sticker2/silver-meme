package calculator

import (
	"DistributedCalc/pkg/errors"
	"fmt"
	"net/http"
)

func NewInvalidExpressionError() *errors.AppError {
	return &errors.AppError{Code: http.StatusUnprocessableEntity, Message: "invalid expression"}
}

func NewDivisionByZeroError() *errors.AppError {
	return &errors.AppError{Code: http.StatusUnprocessableEntity, Message: "division by zero"}
}

func NewInvalidTokenError(token string) *errors.AppError {
	return &errors.AppError{Code: http.StatusUnprocessableEntity, Message: fmt.Sprintf("invalid token: %s", token)}
}

func NewInvalidOperatorError(op string) *errors.AppError {
	return &errors.AppError{Code: http.StatusUnprocessableEntity, Message: fmt.Sprintf("invalid operator: %s", op)}
}
