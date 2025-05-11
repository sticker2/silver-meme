package orchestrator

import (
	"DistributedCalc/internal/grpc"
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/logger"
	"context"
	"strconv"
	"testing"
)

func TestOrchestrator_ProcessExpression(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	orch := NewOrchestrator(dbConn, logr)
	clientMock := &grpc.ClientMock{
		CalculateFunc: func(ctx context.Context, expr string) (*grpc.CalcResponse, error) {
			tokens := Tokenize(expr)
			ops := []string{}
			vals := []float64{}

			for _, token := range tokens {
				if IsNumber(token) {
					num, _ := strconv.ParseFloat(token, 64)
					vals = append(vals, num)
				} else if IsOperator(token) {
					if len(vals) < 2 {
						return &grpc.CalcResponse{Error: "invalid expression"}, nil
					}
					b := vals[len(vals)-1]
					vals = vals[:len(vals)-1]
					a := vals[len(vals)-1]
					vals = vals[:len(vals)-1]
					var result float64
					switch token {
					case "+":
						result = a + b
					case "-":
						result = a - b
					case "*":
						result = a * b
					case "/":
						if b == 0 {
							return &grpc.CalcResponse{Error: "division by zero"}, nil
						}
						result = a / b
					}
					vals = append(vals, result)
				}
			}
			if len(vals) != 1 {
				return &grpc.CalcResponse{Error: "invalid expression"}, nil
			}
			return &grpc.CalcResponse{Result: vals[0]}, nil
		},
	}
	orch.SetGRPCClient(clientMock)

	tests := []struct {
		name     string
		expr     string
		exprID   int64
		expected float64
		err      error
	}{
		{
			name:     "Simple addition",
			expr:     "2+2",
			exprID:   1,
			expected: 4,
			err:      nil,
		},
		{
			name:     "Complex expression",
			expr:     "2+2*2",
			exprID:   2,
			expected: 6,
			err:      nil,
		},
		{
			name:   "Invalid expression",
			expr:   "2/0",
			exprID: 3,
			err:    NewInvalidExpressionError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := orch.ProcessExpression(context.Background(), tt.expr, tt.exprID)
			if tt.err != nil {
				if err == nil || err.Error() != tt.err.Error() {
					t.Errorf("Expected error %v, got %v", tt.err, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}
