package calculator

import (
	"os"
	"testing"
)

func TestCalculator_Evaluate(t *testing.T) {
	calc := NewCalculator()
	tests := []struct {
		name     string
		expr     string
		expected float64
		err      error
	}{
		{
			name:     "Simple addition",
			expr:     "2+2",
			expected: 4,
			err:      nil,
		},
		{
			name:     "Complex expression",
			expr:     "2+2*2",
			expected: 6,
			err:      nil,
		},
		{
			name:     "Parentheses",
			expr:     "2+2)*2",
			expected: 0,
			err:      NewInvalidExpressionError(),
		},
		{
			name: "Division by zero",
			expr: "2/0",
			err:  NewDivisionByZeroError(),
		},
		{
			name: "Invalid token",
			expr: "2+a",
			err:  NewInvalidTokenError("a"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Evaluate(tt.expr)
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

func TestCalculator_ComputeTask(t *testing.T) {
	calc := NewCalculator()
	os.Setenv("TIME_ADDITION_MS", "10")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "20")

	tests := []struct {
		name     string
		arg1     float64
		arg2     float64
		op       string
		expected float64
		err      error
	}{
		{
			name:     "Addition",
			arg1:     2,
			arg2:     3,
			op:       "+",
			expected: 5,
			err:      nil,
		},
		{
			name:     "Multiplication",
			arg1:     2,
			arg2:     3,
			op:       "*",
			expected: 6,
			err:      nil,
		},
		{
			name: "Division by zero",
			arg1: 2,
			arg2: 0,
			op:   "/",
			err:  NewDivisionByZeroError(),
		},
		{
			name: "Invalid operator",
			arg1: 2,
			arg2: 3,
			op:   "^",
			err:  NewInvalidOperatorError("^"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.ComputeTask(tt.arg1, tt.arg2, tt.op)
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
