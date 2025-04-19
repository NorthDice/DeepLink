package main

import (
	"kafka-service/internal/config"
	"kafka-service/internal/handler"
	"kafka-service/internal/services/consumer"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

var (
	consumerGroup = "post-consumer"
)
var address = []string{"localhost:29092"}

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application")

	log.Info(cfg.Env)

	/*mongoClient, err := mongodb.New(cfg.Database.MongoDB.Uri, cfg.Database.MongoDB.Name)
	if err != nil {
		log.Error("failed to connect to mongodb", "error", err)
	}
	defer mongoClient.Close()
	*/
	h := handler.NewHandler(*log)
	c, err := consumer.NewConsumer(h, address, consumerGroup, "post_created.events", log)
	if err != nil {
		log.Error("failed to create consumer", "error", err)
		os.Exit(1)
	}

	go func() {
		c.Start()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	c.Close()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
