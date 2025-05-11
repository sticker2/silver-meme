package calculator

import (
	"DistributedCalc/internal/orchestrator"
	"os"
	"strconv"
	"strings"
	"time"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Evaluate(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	if !orchestrator.IsValidExpression(expr) {
		return 0, NewInvalidExpressionError()
	}

	tokens := orchestrator.Tokenize(expr)
	ops := []string{}
	vals := []float64{}

	for _, token := range tokens {
		if orchestrator.IsNumber(token) {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, NewInvalidTokenError(token)
			}
			vals = append(vals, num)
		} else if token == "(" {
			ops = append(ops, token)
		} else if token == ")" {
			for len(ops) > 0 && ops[len(ops)-1] != "(" {
				if err := c.applyOp(&vals, &ops); err != nil {
					return 0, err
				}
			}
			if len(ops) == 0 {
				return 0, NewInvalidExpressionError()
			}
			ops = ops[:len(ops)-1]
		} else if orchestrator.IsOperator(token) {
			for len(ops) > 0 && ops[len(ops)-1] != "(" && orchestrator.Precedence(ops[len(ops)-1]) >= orchestrator.Precedence(token) {
				if err := c.applyOp(&vals, &ops); err != nil {
					return 0, err
				}
			}
			ops = append(ops, token)
		} else {
			return 0, NewInvalidTokenError(token)
		}
	}

	for len(ops) > 0 {
		if ops[len(ops)-1] == "(" {
			return 0, NewInvalidExpressionError()
		}
		if err := c.applyOp(&vals, &ops); err != nil {
			return 0, err
		}
	}

	if len(vals) != 1 {
		return 0, NewInvalidExpressionError()
	}

	return vals[0], nil
}

func (c *Calculator) ComputeTask(arg1, arg2 float64, op string) (float64, error) {
	var operationTime time.Duration
	switch op {
	case "+":
		operationTime = getOperationTime("TIME_ADDITION_MS")
	case "-":
		operationTime = getOperationTime("TIME_SUBTRACTION_MS")
	case "*":
		operationTime = getOperationTime("TIME_MULTIPLICATIONS_MS")
	case "/":
		operationTime = getOperationTime("TIME_DIVISIONS_MS")
	default:
		return 0, NewInvalidOperatorError(op)
	}

	time.Sleep(operationTime)

	switch op {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, NewDivisionByZeroError()
		}
		return arg1 / arg2, nil
	default:
		return 0, NewInvalidOperatorError(op)
	}
}

func (c *Calculator) applyOp(vals *[]float64, ops *[]string) error {
	if len(*vals) < 2 || len(*ops) == 0 {
		return NewInvalidExpressionError()
	}
	op := (*ops)[len(*ops)-1]
	*ops = (*ops)[:len(*ops)-1]
	b := (*vals)[len(*vals)-1]
	*vals = (*vals)[:len(*vals)-1]
	a := (*vals)[len(*vals)-1]
	*vals = (*vals)[:len(*vals)-1]

	result, err := c.ComputeTask(a, b, op)
	if err != nil {
		return err
	}
	*vals = append(*vals, result)
	return nil
}

func getOperationTime(envVar string) time.Duration {
	ms, _ := strconv.Atoi(os.Getenv(envVar))
	if ms <= 0 {
		ms = 100 // Default 100ms
	}
	return time.Duration(ms) * time.Millisecond
}
