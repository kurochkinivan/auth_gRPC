package app

import (
	"context"
	"log/slog"

	grpcapp "github.com/kurochkinivan/auth/internal/app/grpc"
	pgapp "github.com/kurochkinivan/auth/internal/app/pg"
	"github.com/kurochkinivan/auth/internal/config"
	"github.com/kurochkinivan/auth/internal/usecase/auth"
	"github.com/kurochkinivan/auth/internal/usecase/repository/pg"
)

type App struct {
	log           *slog.Logger
	GRPCApp       *grpcapp.App
	PostgreSQLApp *pgapp.App
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	pgApp := pgapp.New(ctx, log, cfg.PostgreSQL)

	repository := pg.New(pgApp.Pool)

	authService := auth.New(log, repository, repository, cfg.Secret, cfg.TokenTTL)

	gRPCApp := grpcapp.New(log, cfg.GRPC, authService)

	return &App{
		GRPCApp: gRPCApp,
		PostgreSQLApp: pgApp,
		log:     log,
	}
}
