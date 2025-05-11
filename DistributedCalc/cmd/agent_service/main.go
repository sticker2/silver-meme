package main

import (
	"DistributedCalc/internal/calculator"
	"DistributedCalc/internal/grpc"
	"DistributedCalc/internal/tasks"
	"DistributedCalc/pkg/logger"
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	logr := logger.NewLogger()
	calc := calculator.NewCalculator()
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logr.Error("Failed to listen: %v", err)
		os.Exit(1)
	}

	server := grpc.NewServer()
	grpc.RegisterCalcServiceServer(server, grpc.NewServer(logr))

	go func() {
		logr.Info("Starting agent_service on :50051")
		if err := server.Serve(listener); err != nil {
			logr.Error("Server failed: %v", err)
			os.Exit(1)
		}
	}()

	computingPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if computingPower <= 0 {
		computingPower = 4
	}

	taskClient := tasks.NewTaskClient("http://localhost:8080", logr)
	for i := 0; i < computingPower; i++ {
		go taskClient.RunWorker(context.Background(), calc)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Info("Shutting down server...")
	server.GracefulStop()
	logr.Info("Server stopped")
}
