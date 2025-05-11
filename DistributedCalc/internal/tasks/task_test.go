package tasks

import (
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/logger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskService_GetTaskHandler(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	taskService := NewTaskService(dbConn, logr)

	dbConn.CreateUser("testuser", "hashedpassword")
	exprID, err := dbConn.SaveExpression("testuser", "2+2", "pending")
	if err != nil {
		t.Fatalf("Failed to save expression: %v", err)
	}

	taskID, err := dbConn.SaveTask(exprID, 2, 2, "+", 100)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	req := httptest.NewRequest("GET", "/internal/task", nil)
	rr := httptest.NewRecorder()
	taskService.GetTaskHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	taskResp, ok := resp["task"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected task in response")
	}
	if taskResp["id"].(float64) != float64(taskID) {
		t.Errorf("Expected task ID %d, got %v", taskID, taskResp["id"])
	}
}

func TestTaskService_SubmitTaskResult(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	taskService := NewTaskService(dbConn, logr)

	dbConn.CreateUser("testuser", "hashedpassword")
	exprID, err := dbConn.SaveExpression("testuser", "2+2", "pending")
	if err != nil {
		t.Fatalf("Failed to save expression: %v", err)
	}

	taskID, err := dbConn.SaveTask(exprID, 2, 2, "+", 100)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	resultBody := `{"id": ` + string(rune(taskID)) + `, "result": 4}`
	req := httptest.NewRequest("POST", "/internal/task", bytes.NewBufferString(resultBody))
	rr := httptest.NewRecorder()
	taskService.GetTaskHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
