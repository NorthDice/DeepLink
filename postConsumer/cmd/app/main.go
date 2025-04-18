package main

import (
	"context"
	"kafka-service/internal/config"
	"kafka-service/internal/services/consumer"
	"kafka-service/internal/storage/mongodb"
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

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application")

	log.Info(cfg.Env)

	mongoClient, err := mongodb.New(cfg.Database.MongoDB.Uri, cfg.Database.MongoDB.Name)
	if err != nil {
		log.Error("failed to connect to mongodb", "error", err)
	}
	defer mongoClient.Close()

	kafkaConsumer, err := consumer.New(log, cfg.Kafka.Brokers, "post-consumer-group", mongoClient, []string{
		"post_created.events",
		"post_deleted.events",
		"post_liked.events",
		"comment_added.events",
	})

	if err != nil {
		log.Error("failed to initialize kafka consumer", "error", err)
	}

	go func() {
		if err := kafkaConsumer.Run(context.Background()); err != nil {
			log.Error("consumer error:", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down...")
	if err := kafkaConsumer.Close(); err != nil {
		log.Error("failed to close consumer:", err)
	}
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
