package integration_tests

import (
	"DistributedCalc/internal/auth"
	"DistributedCalc/internal/calculator"
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/logger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculatorService_Integration(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	authService := auth.NewAuthService(dbConn, logr)
	calcService := calculator.NewCalculatorService(dbConn, logr)

	// Register user
	registerBody := `{"login": "testuser", "password": "password123"}`
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(registerBody))
	rr := httptest.NewRecorder()
	authService.RegisterHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to register user: status %d", rr.Code)
	}

	// Login user
	req = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(registerBody))
	rr = httptest.NewRecorder()
	authService.LoginHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to login user: status %d", rr.Code)
	}

	var loginResp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}
	token := loginResp["token"]

	// Calculate expression
	calcBody := `{"expression": "2+2"}`
	req = httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(calcBody))
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	handler := auth.JWTMiddleware(http.HandlerFunc(calcService.CalculateHandler), authService)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to calculate expression: status %d", rr.Code)
	}

	var calcResp calculator.CalcResponse
	if err := json.NewDecoder(rr.Body).Decode(&calcResp); err != nil {
		t.Fatalf("Failed to decode calc response: %v", err)
	}
	if calcResp.ID <= 0 {
		t.Errorf("Expected positive expression ID, got %d", calcResp.ID)
	}
}
