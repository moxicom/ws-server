package server

import (
	"context"
	"github.com/moxicom/ws-server/api/internal/auth"
	"github.com/moxicom/ws-server/api/internal/config"
	"github.com/moxicom/ws-server/api/internal/storage/redis_storage"
	"github.com/moxicom/ws-server/api/internal/ws"
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
	cfg      config.Config
}

func New(logger *slog.Logger, cfg config.Config) *Server {
	return &Server{
		Shutdown: make(chan struct{}),
		logger:   logger,
		cfg:      cfg,
	}
}

func (s *Server) Run() {
	logger := s.logger.With(slog.String("op", "server.Run"))

	// Connect to redis_storage
	storageService := redis_storage.MustConnect(s.cfg.RedisAddr, s.cfg.RedisPass, s.cfg.RedisDB)

	// Init hub
	hub := ws.NewHub(s.logger, storageService)
	go hub.Run()

	// TODO: Connect to redis_storage
	// rdb := connect_to_redis_mock

	http.HandleFunc("/ws", auth.AuthMiddleware(hub, ws.Upgrade))

	server := &http.Server{
		Addr:    s.cfg.ServerAddr,
		Handler: nil,
	}

	go func() {
		logger.Info("Server listening on " + s.cfg.ServerAddr)
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
