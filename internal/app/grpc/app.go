package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kurochkinivan/auth/internal/config"
	authgrpc "github.com/kurochkinivan/auth/internal/controller/grpc/auth"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       string
	timeout    time.Duration
}

func New(log *slog.Logger, cfg config.GRPCConfig, auth authgrpc.Auth) *App {
	gRPCServer := grpc.NewServer(grpc.ConnectionTimeout(cfg.Timeout))

	validate := validator.New(validator.WithRequiredStructEnabled())

	authgrpc.Register(gRPCServer, validate, auth)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       cfg.Port,
		timeout:    cfg.Timeout,
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
		slog.String("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping grpc server...", slog.String("port", a.port))

	a.gRPCServer.GracefulStop()
}
