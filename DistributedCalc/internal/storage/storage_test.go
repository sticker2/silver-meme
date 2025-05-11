package storage

import (
	"DistributedCalc/pkg/logger"
	"testing"
)

func TestSQLiteDB_CreateUser(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	err = dbConn.CreateUser("testuser", "hashedpassword")
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	err = dbConn.CreateUser("testuser", "anotherpassword")
	if err == nil || err.Error() != NewUserExistsError().Error() {
		t.Error("Expected error for duplicate user, got", err)
	}
}

func TestSQLiteDB_SaveExpression(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	dbConn.CreateUser("testuser", "hashedpassword")
	id, err := dbConn.SaveExpression("testuser", "2+2", "pending")
	if err != nil {
		t.Errorf("Failed to save expression: %v", err)
	}
	if id <= 0 {
		t.Error("Expected positive ID, got", id)
	}
}

func TestSQLiteDB_SaveTask(t *testing.T) {
	logr := logger.NewLogger()
	dbConn, err := NewSQLiteDB(":memory:", logr)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer dbConn.Close()

	dbConn.CreateUser("testuser", "hashedpassword")
	exprID, err := dbConn.SaveExpression("testuser", "2+2", "pending")
	if err != nil {
		t.Fatalf("Failed to save expression: %v", err)
	}

	taskID, err := dbConn.SaveTask(exprID, 2, 2, "+", 100)
	if err != nil {
		t.Errorf("Failed to save task: %v", err)
	}
	if taskID <= 0 {
		t.Error("Expected positive task ID, got", taskID)
	}
}
