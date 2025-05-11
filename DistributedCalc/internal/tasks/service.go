package tasks

import (
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/errors"
	"DistributedCalc/pkg/logger"
	"encoding/json"
	"net/http"
)

type TaskService struct {
	db   *storage.SQLiteDB
	logr *logger.Logger
}

func NewTaskService(db *storage.SQLiteDB, logr *logger.Logger) *TaskService {
	return &TaskService{
		db:   db,
		logr: logr,
	}
}

func (s *TaskService) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := s.db.GetPendingTask()
	if err != nil {
		if err.Error() == "no pending tasks" {
			errors.HandleHTTPError(w, errors.NewNotFoundError("no pending tasks"))
			return
		}
		s.logr.Error("Failed to get pending task: %v", err)
		errors.HandleHTTPError(w, errors.NewInternalError("failed to get task"))
		return
	}

	taskResponse := Task{
		ID:            task.ID,
		ExpressionID:  task.ExpressionID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operator,
		OperationTime: task.Duration,
	}
	json.NewEncoder(w).Encode(taskResponse)
}

func (s *TaskService) SubmitTaskResultHandler(w http.ResponseWriter, r *http.Request) {
	var result TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		s.logr.Error("Failed to decode task result: %v", err)
		errors.HandleHTTPError(w, errors.NewBadRequestError("invalid request body"))
		return
	}

	status := "completed"
	if result.Result == 0 {
		status = "error"
	}

	if err := s.db.UpdateTaskResult(result.ID, result.Result, status); err != nil {
		s.logr.Error("Failed to update task %d: %v", result.ID, err)
		errors.HandleHTTPError(w, errors.NewInternalError("failed to update task"))
		return
	}

	w.WriteHeader(http.StatusOK)
}
