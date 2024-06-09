package main

import (
	"github.com/moxicom/ws-server/internal/server"
	"log/slog"
	"os"
)

func main() {
	log := setupLogger()

	s := server.New(log)
	s.Run(":8080")

}

func setupLogger() *slog.Logger {
	var log *slog.Logger
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}
