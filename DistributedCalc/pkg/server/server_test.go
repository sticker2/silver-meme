package server

import (
	"DistributedCalc/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_AddRoute(t *testing.T) {
	logr := logger.NewLogger()
	srv := NewServer(":0", logr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	srv.AddRoute("/test", handler, "GET")

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %s", rr.Body.String())
	}
}

func TestServer_MethodNotAllowed(t *testing.T) {
	logr := logger.NewLogger()
	srv := NewServer(":0", logr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	srv.AddRoute("/test", handler, "GET")

	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
