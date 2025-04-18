package main

import (
	"fmt"
	"github.com/yourusername/deepLink/post-service/internal/config"
	"github.com/yourusername/deepLink/post-service/internal/services/producer"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var address = []string{"localhost:29092"}

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting app",
		slog.String("env", cfg.Env),
		slog.Any("config", cfg),
	)

	p, err := producer.NewProducer(address, log)
	if err != nil {
		log.Error("error creating producer")
	}

	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("msg %d", i)
		if err = p.Produce(msg, "post_created"); err != nil {
			log.Error("error producing message")
		}
	}
	/*
		uri := cfg.Database.Mongo.Uri

		if uri == "" {
			panic("database url is required")
		}

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		if err != nil {
			panic(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			panic(err)
		}

		application := app.New(log, cfg.GRPC.Port, "fsd")

		go application.GRPCServer.MustRun()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

		<-stop

		application.GRPCServer.Stop()

		log.Info("shutting down")

	*/
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
