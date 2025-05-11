package main

import (
	"DistributedCalc/internal/auth"
	"DistributedCalc/internal/calculator"
	"DistributedCalc/internal/grpc"
	"DistributedCalc/internal/orchestrator"
	"DistributedCalc/internal/storage"
	"DistributedCalc/internal/tasks"
	"DistributedCalc/pkg/logger"
	"DistributedCalc/pkg/server"
	"context"
	"net/http"
	"time"
)

func main() {
	logr := logger.NewLogger()
	dbConn, err := storage.NewSQLiteDB("calc.db", logr)
	if err != nil {
		logr.Error("Failed to init DB: %v", err)
		return
	}
	defer dbConn.Close()

	authService := auth.NewAuthService(dbConn, logr)
	calcService := calculator.NewCalculatorService(dbConn, logr)
	taskService := tasks.NewTaskService(dbConn, logr)
	orch := orchestrator.NewOrchestrator(dbConn, logr)
	client, err := grpc.NewClient(":50051", logr)
	if err != nil {
		logr.Error("Failed to init gRPC client: %v", err)
		return
	}
	orch.SetGRPCClient(client)

	go processExpressions(dbConn, orch, logr)

	srv := server.NewServer(":8080", logr)
	srv.AddRoute("/api/v1/register", http.HandlerFunc(authService.RegisterHandler), "POST")
	srv.AddRoute("/api/v1/login", http.HandlerFunc(authService.LoginHandler), "POST")
	srv.AddRoute("/api/v1/calculate", authService.JWTMiddleware(http.HandlerFunc(calcService.CalculateHandler), authService), "POST")
	srv.AddRoute("/api/v1/expressions", authService.JWTMiddleware(http.HandlerFunc(calcService.ListExpressionsHandler), authService), "GET")
	srv.AddRoute("/api/v1/expression", authService.JWTMiddleware(http.HandlerFunc(calcService.GetExpressionHandler), authService), "GET")
	srv.AddRoute("/api/v1/task", http.HandlerFunc(taskService.GetTaskHandler), "GET")
	srv.AddRoute("/api/v1/task/result", http.HandlerFunc(taskService.SubmitTaskResultHandler), "POST")

	if err := srv.Run(); err != nil {
		logr.Error("Server failed: %v", err)
	}
}

func processExpressions(db *storage.SQLiteDB, orch *orchestrator.Orchestrator, logr *logger.Logger) {
	for {
		exprs, err := db.GetPendingExpressions()
		if err != nil {
			logr.Error("Failed to get pending expressions: %v", err)
			continue
		}
		for _, expr := range exprs {
			go func(expr storage.Expression) {
				result, err := orch.ProcessExpression(context.Background(), expr.Expression, expr.ID)
				if err != nil {
					logr.Error("Failed to process expression %d: %v", expr.ID, err)
					db.UpdateExpression(expr.ID, 0, "error")
					return
				}
				db.UpdateExpression(expr.ID, result, "completed")
			}(expr)
		}
		time.Sleep(1 * time.Second)
	}
}
