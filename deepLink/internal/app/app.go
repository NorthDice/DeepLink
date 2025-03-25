package app

import (
	grpcapp "github.com/NorthDice/DeepLink/internal/app/grpc"
	"github.com/NorthDice/DeepLink/internal/services/auth"
	mypostgres "github.com/NorthDice/DeepLink/internal/storage/postgres"

	"log/slog"
	"time"
)

type App struct {
	GRPCSrv  *grpcapp.App
	Postgres *mypostgres.Storage
}

func New(
	log *slog.Logger,
	grpcPort int,
	storage *mypostgres.Storage,
	tokenTTL time.Duration,
) *App {

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
