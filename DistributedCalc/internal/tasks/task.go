package tasks

import (
	"DistributedCalc/internal/calculator"
	"DistributedCalc/pkg/logger"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Task struct {
	ID            int64
	ExpressionID  int64
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime int
}

func (t *Task) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":             t.ID,
		"arg1":           t.Arg1,
		"arg2":           t.Arg2,
		"operation":      t.Operation,
		"operation_time": t.OperationTime,
	}
}

type TaskResult struct {
	ID     int64   `json:"id"`
	Result float64 `json:"result"`
}

type TaskClient struct {
	addr   string
	client *http.Client
	logr   *logger.Logger
}

func NewTaskClient(addr string, logr *logger.Logger) *TaskClient {
	return &TaskClient{
		addr:   addr,
		client: &http.Client{Timeout: 10 * time.Second},
		logr:   logr,
	}
}

func (c *TaskClient) RunWorker(ctx context.Context, calc *calculator.Calculator) {
	for {
		select {
		case <-ctx.Done():
			c.logr.Info("Task worker stopped")
			return
		default:
			task, err := c.fetchTask()
			if err != nil {
				if err.Error() == NewTaskNotFoundError().Error() {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				c.logr.Error("Failed to fetch task: %v", err)
				continue
			}

			result, err := calc.ComputeTask(task.Arg1, task.Arg2, task.Operation)
			if err != nil {
				c.logr.Error("Failed to compute task %d: %v", task.ID, err)
				c.submitTaskResult(task.ID, 0, "error")
				continue
			}

			if err := c.submitTaskResult(task.ID, result, "completed"); err != nil {
				c.logr.Error("Failed to submit task %d result: %v", task.ID, err)
			}
		}
	}
}

func (c *TaskClient) fetchTask() (Task, error) {
	resp, err := c.client.Get(c.addr + "/api/v1/task")
	if err != nil {
		c.logr.Error("Failed to fetch task from %s: %v", c.addr, err)
		return Task{}, NewTaskFetchError(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Task{}, NewTaskNotFoundError()
	}
	if resp.StatusCode != http.StatusOK {
		c.logr.Error("Unexpected status code: %d", resp.StatusCode)
		return Task{}, NewTaskFetchError("unexpected response status")
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		c.logr.Error("Failed to decode task: %v", err)
		return Task{}, NewInvalidTaskError("invalid task data")
	}
	return task, nil
}

func (c *TaskClient) submitTaskResult(taskID int64, result float64, status string) error {
	taskResult := TaskResult{ID: taskID, Result: result}
	body, err := json.Marshal(taskResult)
	if err != nil {
		c.logr.Error("Failed to marshal task result: %v", err)
		return NewTaskSubmitError()
	}

	resp, err := c.client.Post(c.addr+"/api/v1/task/result", "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logr.Error("Failed to submit task result to %s: %v", c.addr, err)
		return NewTaskSubmitError()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logr.Error("Unexpected status code: %d", resp.StatusCode)
		return NewTaskSubmitError()
	}
	return nil
}
