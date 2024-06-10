package main

import (
	"github.com/joho/godotenv"
	"github.com/moxicom/ws-server/api/internal/config"
	"github.com/moxicom/ws-server/api/internal/server"
	"log"
	"log/slog"
	"os"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	log := setupLogger()

	config := config.NewConfig()

	s := server.New(log, config)
	s.Run()

}

func setupLogger() *slog.Logger {
	var log *slog.Logger
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}
