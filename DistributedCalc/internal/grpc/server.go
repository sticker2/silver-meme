package grpc

import (
	"DistributedCalc/pkg/logger"
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	UnimplementedCalcServiceServer
	logr *logger.Logger
}

func NewServer(logr *logger.Logger) *Server {
	return &Server{logr: logr}
}

func (s *Server) Calculate(ctx context.Context, req *CalcRequest) (*CalcResponse, error) {
	s.logr.Info("Received gRPC request: %s", req.Expression)
	tokens := strings.Split(req.Expression, "")
	if len(tokens) != 3 {
		s.logr.Error("Invalid gRPC expression format: %s", req.Expression)
		return &CalcResponse{Error: "invalid expression"}, nil
	}

	arg1, err := strconv.ParseFloat(tokens[0], 64)
	if err != nil {
		s.logr.Error("Invalid first argument: %s", tokens[0])
		return &CalcResponse{Error: "invalid argument"}, nil
	}
	op := tokens[1]
	arg2, err := strconv.ParseFloat(tokens[2], 64)
	if err != nil {
		s.logr.Error("Invalid second argument: %s", tokens[2])
		return &CalcResponse{Error: "invalid argument"}, nil
	}

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
		s.logr.Error("Invalid operator: %s", op)
		return &CalcResponse{Error: "invalid operator"}, nil
	}

	time.Sleep(operationTime)

	switch op {
	case "+":
		return &CalcResponse{Result: arg1 + arg2}, nil
	case "-":
		return &CalcResponse{Result: arg1 - arg2}, nil
	case "*":
		return &CalcResponse{Result: arg1 * arg2}, nil
	case "/":
		if arg2 == 0 {
			s.logr.Error("Division by zero")
			return &CalcResponse{Error: "division by zero"}, nil
		}
		return &CalcResponse{Result: arg1 / arg2}, nil
	default:
		s.logr.Error("Invalid operator: %s", op)
		return &CalcResponse{Error: "invalid operator"}, nil
	}
}

func getOperationTime(envVar string) time.Duration {
	ms, _ := strconv.Atoi(os.Getenv(envVar))
	if ms <= 0 {
		ms = 100
	}
	return time.Duration(ms) * time.Millisecond
}
