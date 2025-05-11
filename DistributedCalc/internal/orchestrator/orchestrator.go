package orchestrator

import (
	"DistributedCalc/internal/grpc"
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/logger"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Orchestrator struct {
	db     *storage.SQLiteDB
	logr   *logger.Logger
	client *grpc.Client
}

func NewOrchestrator(db *storage.SQLiteDB, logr *logger.Logger) *Orchestrator {
	return &Orchestrator{db: db, logr: logr}
}

func (o *Orchestrator) SetGRPCClient(client *grpc.Client) {
	o.client = client
}

func (o *Orchestrator) ProcessExpression(ctx context.Context, expr string, exprID int64) (float64, error) {
	o.logr.Info("Processing expression %s (ID: %d)", expr, exprID)
	tokens := Tokenize(expr)
	ops := []string{}
	vals := []float64{}
	var mu sync.Mutex

	taskChan := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, token := range tokens {
		if IsNumber(token) {
			num, err := parseNumber(token)
			if err != nil {
				return 0, err
			}
			mu.Lock()
			vals = append(vals, num)
			mu.Unlock()
		} else if token == "(" {
			mu.Lock()
			ops = append(ops, token)
			mu.Unlock()
		} else if token == ")" {
			mu.Lock()
			for len(ops) > 0 && ops[len(ops)-1] != "(" {
				op := ops[len(ops)-1]
				ops = ops[:len(ops)-1]
				if len(vals) < 2 {
					mu.Unlock()
					return 0, NewInvalidExpressionError()
				}
				b := vals[len(vals)-1]
				vals = vals[:len(vals)-1]
				a := vals[len(vals)-1]
				vals = vals[:len(vals)-1]
				mu.Unlock()

				operationTime := getOperationTime(op)
				taskID, err := o.db.SaveTask(exprID, a, b, op, operationTime)
				if err != nil {
					return 0, NewTaskDistributionError("failed to save task")
				}

				wg.Add(1)
				taskChan <- struct{}{}
				go func(a, b float64, op string, taskID int64) {
					defer wg.Done()
					defer func() { <-taskChan }()
					resp, err := o.client.Calculate(ctx, formatTask(a, b, op))
					if err != nil {
						o.logr.Error("gRPC calculation failed for task %d: %v", taskID, err)
						o.db.UpdateTaskResult(taskID, 0, "error")
						return
					}
					if resp.Error != "" {
						o.logr.Error("Calculation error for task %d: %s", taskID, resp.Error)
						o.db.UpdateTaskResult(taskID, 0, "error")
						return
					}
					mu.Lock()
					vals = append(vals, resp.Result)
					mu.Unlock()
					o.db.UpdateTaskResult(taskID, resp.Result, "completed")
				}(a, b, op, taskID)
			}
			if len(ops) == 0 {
				mu.Unlock()
				return 0, NewInvalidExpressionError()
			}
			ops = ops[:len(ops)-1]
			mu.Unlock()
		} else if IsOperator(token) {
			mu.Lock()
			for len(ops) > 0 && ops[len(ops)-1] != "(" && Precedence(ops[len(ops)-1]) >= Precedence(token) {
				op := ops[len(ops)-1]
				ops = ops[:len(ops)-1]
				if len(vals) < 2 {
					mu.Unlock()
					return 0, NewInvalidExpressionError()
				}
				b := vals[len(vals)-1]
				vals = vals[:len(vals)-1]
				a := vals[len(vals)-1]
				vals = vals[:len(vals)-1]
				mu.Unlock()

				operationTime := getOperationTime(op)
				taskID, err := o.db.SaveTask(exprID, a, b, op, operationTime)
				if err != nil {
					return 0, NewTaskDistributionError("failed to save task")
				}

				wg.Add(1)
				taskChan <- struct{}{}
				go func(a, b float64, op string, taskID int64) {
					defer wg.Done()
					defer func() { <-taskChan }()
					resp, err := o.client.Calculate(ctx, formatTask(a, b, op))
					if err != nil {
						o.logr.Error("gRPC calculation failed for task %d: %v", taskID, err)
						o.db.UpdateTaskResult(taskID, 0, "error")
						return
					}
					if resp.Error != "" {
						o.logr.Error("Calculation error for task %d: %s", taskID, resp.Error)
						o.db.UpdateTaskResult(taskID, 0, "error")
						return
					}
					mu.Lock()
					vals = append(vals, resp.Result)
					mu.Unlock()
					o.db.UpdateTaskResult(taskID, resp.Result, "completed")
				}(a, b, op, taskID)
			}
			ops = append(ops, token)
			mu.Unlock()
		} else {
			return 0, NewInvalidExpressionError()
		}
	}

	mu.Lock()
	for len(ops) > 0 {
		if ops[len(ops)-1] == "(" {
			mu.Unlock()
			return 0, NewInvalidExpressionError()
		}
		op := ops[len(ops)-1]
		ops = ops[:len(ops)-1]
		if len(vals) < 2 {
			mu.Unlock()
			return 0, NewInvalidExpressionError()
		}
		b := vals[len(vals)-1]
		vals = vals[:len(vals)-1]
		a := vals[len(vals)-1]
		vals = vals[:len(vals)-1]
		mu.Unlock()

		operationTime := getOperationTime(op)
		taskID, err := o.db.SaveTask(exprID, a, b, op, operationTime)
		if err != nil {
			return 0, NewTaskDistributionError("failed to save task")
		}

		wg.Add(1)
		taskChan <- struct{}{}
		go func(a, b float64, op string, taskID int64) {
			defer wg.Done()
			defer func() { <-taskChan }()
			resp, err := o.client.Calculate(ctx, formatTask(a, b, op))
			if err != nil {
				o.logr.Error("gRPC calculation failed for task %d: %v", taskID, err)
				o.db.UpdateTaskResult(taskID, 0, "error")
				return
			}
			if resp.Error != "" {
				o.logr.Error("Calculation error for task %d: %s", taskID, resp.Error)
				o.db.UpdateTaskResult(taskID, 0, "error")
				return
			}
			mu.Lock()
			vals = append(vals, resp.Result)
			mu.Unlock()
			o.db.UpdateTaskResult(taskID, resp.Result, "completed")
		}(a, b, op, taskID)
	}
	mu.Unlock()

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(vals) != 1 {
		return 0, NewInvalidExpressionError()
	}

	return vals[0], nil
}

func Tokenize(expr string) []string {
	var tokens []string
	var num strings.Builder
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if isDigit(ch) || ch == '.' {
			num.WriteByte(ch)
		} else {
			if num.Len() > 0 {
				tokens = append(tokens, num.String())
				num.Reset()
			}
			if IsOperator(string(ch)) || ch == '(' || ch == ')' {
				tokens = append(tokens, string(ch))
			}
		}
	}
	if num.Len() > 0 {
		tokens = append(tokens, num.String())
	}
	return tokens
}

func parseNumber(token string) (float64, error) {
	num, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return 0, NewInvalidExpressionError()
	}
	return num, nil
}

func getOperationTime(op string) int {
	var envVar string
	switch op {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATIONS_MS"
	case "/":
		envVar = "TIME_DIVISIONS_MS"
	default:
		return 100
	}
	ms, _ := strconv.Atoi(os.Getenv(envVar))
	if ms <= 0 {
		ms = 100
	}
	return ms
}

func formatTask(a, b float64, op string) string {
	return fmt.Sprintf("%f%s%f", a, op, b)
}

func IsValidExpression(expr string) bool {
	allowed := "0123456789+-*/()."
	for _, ch := range expr {
		if !strings.ContainsRune(allowed, ch) {
			return false
		}
	}
	return true
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func IsNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

func IsOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

func Precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}
