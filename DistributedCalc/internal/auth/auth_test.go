package auth

import (
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/logger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_RegisterHandler(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	authService := NewAuthService(dbConn, logr)

	tests := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "Valid registration",
			body:       `{"login": "testuser", "password": "password123"}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "Invalid JSON",
			body:       `{"login": "testuser", "password": }`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Duplicate user",
			body:       `{"login": "testuser", "password": "password123"}`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()
			authService.RegisterHandler(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, rr.Code)
			}
		})
	}
}

func TestAuthService_LoginHandler(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	authService := NewAuthService(dbConn, logr)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	dbConn.CreateUser("testuser", string(hashedPassword))

	tests := []struct {
		name       string
		body       string
		statusCode int
		hasToken   bool
	}{
		{
			name:       "Valid login",
			body:       `{"login": "testuser", "password": "password123"}`,
			statusCode: http.StatusOK,
			hasToken:   true,
		},
		{
			name:       "Invalid password",
			body:       `{"login": "testuser", "password": "wrong"}`,
			statusCode: http.StatusUnauthorized,
			hasToken:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()
			authService.LoginHandler(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, rr.Code)
			}
			if tt.hasToken {
				var resp map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if _, ok := resp["token"]; !ok {
					t.Error("Expected token in response")
				}
			}
		})
	}
}
