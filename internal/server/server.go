package server

import (
	"context"
	"fmt"
	"github.com/moxicom/ws-server/internal/ws"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Shutdown chan struct{}
}

func New() *Server {
	return &Server{
		Shutdown: make(chan struct{}),
	}
}

func (s *Server) Run(addr string) {
	hub := ws.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.Upgrade(hub, w, r)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: nil,
	}

	log.Println("Server started")
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("%s", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	fmt.Println("Shutting down...")

	// Shutdown WebSocket hub
	hub.Shutdown()

	// Gracefully shutdown the server without interrupting any active connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}
