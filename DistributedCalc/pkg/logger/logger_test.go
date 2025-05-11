package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{logger: log.New(&buf, "DistributedCalc: ", log.LstdFlags)}
	logger.Info("Test message: %s", "info")

	output := buf.String()
	if !strings.Contains(output, "[INFO] Test message: info") {
		t.Errorf("Expected log to contain '[INFO] Test message: info', got %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{logger: log.New(&buf, "DistributedCalc: ", log.LstdFlags)}
	logger.Error("Test message: %s", "error")

	output := buf.String()
	if !strings.Contains(output, "[ERROR] Test message: error") {
		t.Errorf("Expected log to contain '[ERROR] Test message: error', got %s", output)
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{logger: log.New(&buf, "DistributedCalc: ", log.LstdFlags)}
	logger.Debug("Test message: %s", "debug")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG] Test message: debug") {
		t.Errorf("Expected log to contain '[DEBUG] Test message: debug', got %s", output)
	}
}
