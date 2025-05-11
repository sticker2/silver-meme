package server

import (
	"DistributedCalc/pkg/errors"
	"DistributedCalc/pkg/logger"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	addr   string
	router *mux.Router
	logr   *logger.Logger
}

func NewServer(addr string, logr *logger.Logger) *Server {
	return &Server{
		addr:   addr,
		router: mux.NewRouter(),
		logr:   logr,
	}
}

func (s *Server) AddRoute(path string, handler http.Handler, methods ...string) {
	s.router.Handle(path, handler).Methods(methods...)
	s.logr.Info("Added route: %s [%s]", path, methods)
}

func (s *Server) Run() error {
	s.router.Use(s.loggingMiddleware)
	s.logr.Info("Starting server on %s", s.addr)
	if err := http.ListenAndServe(s.addr, s.router); err != nil {
		s.logr.Error("Server failed to start: %v", err)
		return errors.NewInternalError("server failed to start")
	}
	return nil
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logr.Info("Received %s request to %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
