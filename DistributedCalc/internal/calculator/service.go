package calculator

import (
	"DistributedCalc/internal/auth"
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/errors"
	"DistributedCalc/pkg/logger"
	"encoding/json"
	"net/http"
	"strconv"
)

type CalculatorService struct {
	db   *storage.SQLiteDB
	logr *logger.Logger
}

func NewCalculatorService(db *storage.SQLiteDB, logr *logger.Logger) *CalculatorService {
	return &CalculatorService{db: db, logr: logr}
}

type CalcRequest struct {
	Expression string `json:"expression"`
}

type CalcResponse struct {
	ID int64 `json:"id"`
}

func (s *CalculatorService) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req CalcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logr.Error("Failed to decode request: %v", err)
		errors.HandleHTTPError(w, errors.NewBadRequestError("invalid request body"))
		return
	}

	userID, ok := r.Context().Value(auth.UserIDKey).(int64)
	if !ok {
		s.logr.Error("User ID not found in context")
		errors.HandleHTTPError(w, errors.NewInternalError("user not authenticated"))
		return
	}

	id, err := s.db.SaveExpression(userID, req.Expression)
	if err != nil {
		s.logr.Error("Failed to save expression: %v", err)
		errors.HandleHTTPError(w, errors.NewInternalError("failed to save expression"))
		return
	}

	s.logr.Info("Expression saved with ID %d for user %d", id, userID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CalcResponse{ID: id})
}

func (s *CalculatorService) ListExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int64)
	if !ok {
		s.logr.Error("User ID not found in context")
		errors.HandleHTTPError(w, errors.NewInternalError("user not authenticated"))
		return
	}

	exprs, err := s.db.GetUserExpressions(userID)
	if err != nil {
		s.logr.Error("Failed to get expressions: %v", err)
		errors.HandleHTTPError(w, errors.NewInternalError("failed to get expressions"))
		return
	}

	json.NewEncoder(w).Encode(exprs)
}

func (s *CalculatorService) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int64)
	if !ok {
		s.logr.Error("User ID not found in context")
		errors.HandleHTTPError(w, errors.NewInternalError("user not authenticated"))
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.logr.Error("Invalid expression ID: %v", err)
		errors.HandleHTTPError(w, errors.NewBadRequestError("invalid expression ID"))
		return
	}

	expr, err := s.db.GetExpression(id, userID)
	if err != nil {
		s.logr.Error("Failed to get expression %d: %v", id, err)
		if err.Error() == "expression not found" {
			errors.HandleHTTPError(w, errors.NewNotFoundError("expression not found"))
		} else {
			errors.HandleHTTPError(w, errors.NewInternalError("failed to get expression"))
		}
		return
	}

	json.NewEncoder(w).Encode(expr)
}
