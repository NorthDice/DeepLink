package app

import (
	"github.com/yourusername/deepLink/post-service/internal/app/grpcapp"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagepath string,
) *App {
	//TODO : INIT STORAGE
	//TODO : INIT POST SERVICE

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
