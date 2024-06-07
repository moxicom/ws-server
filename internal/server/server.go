package server

import (
	"context"
	"github.com/moxicom/ws-server/internal/ws"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Shutdown chan struct{}
	logger   *slog.Logger
}

func New(logger *slog.Logger) *Server {
	return &Server{
		Shutdown: make(chan struct{}),
		logger:   logger,
	}
}

func (s *Server) Run(addr string) {
	logger := s.logger.With(slog.String("op", "server.Run"))

	hub := ws.NewHub(s.logger)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.Upgrade(hub, w, r)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: nil,
	}

	go func() {
		logger.Info("Server listening on " + addr)
		err := server.ListenAndServe()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down server...")

	// Shutdown WebSocket hub
	hub.Shutdown()

	// Gracefully shutdown the server without interrupting any active connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server Shutdown error: %v", err.Error())
		panic(err)
	}

	logger.Info("Server gracefully stopped")
}
