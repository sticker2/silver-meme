package orchestrator

import (
	"DistributedCalc/pkg/errors"
	"fmt"
)

func NewInvalidExpressionError() error {
	return errors.NewBadRequestError("invalid expression")
}

func NewTaskDistributionError(msg string) error {
	return errors.NewInternalError(fmt.Sprintf("task distribution error: %s", msg))
}
