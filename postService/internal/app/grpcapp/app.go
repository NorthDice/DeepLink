package grpcapp

import (
	"fmt"
	"github.com/yourusername/deepLink/post-service/internal/grpc/post"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	post.Register(gRPCServer)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {

	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	return nil
}

func (a *App) Stop() {

	const op = "grpcapp.Stop"

	a.log.With(
		slog.String("op", op)).
		Info("stopping grpc server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
