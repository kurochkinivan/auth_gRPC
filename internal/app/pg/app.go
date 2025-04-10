package pgapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/auth/internal/config"
	"github.com/kurochkinivan/auth/internal/lib/sl"
	"github.com/kurochkinivan/auth/pkg/pgclient"
)

type App struct {
	Pool     *pgxpool.Pool
	log      *slog.Logger
	host     string
	port     string
	username string
	password string
	db       string
}

func New(ctx context.Context, log *slog.Logger, cfg config.PostgreSQLConfig) *App {
	pool, err := pgclient.NewClient(ctx, cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	if err != nil {
		panic("pgapp.New: " + err.Error())
	}

	return &App{
		log:      log,
		Pool:     pool,
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		db:       cfg.DB,
	}
}

func (a *App) MustRun(ctx context.Context, maxAttempts int, delay time.Duration) {
	if err := a.Run(ctx, maxAttempts, delay); err != nil {
		panic(err)
	}
}

func (a *App) Run(ctx context.Context, maxAttempts int, delay time.Duration) error {
	const op = "pgapp.Run()"
	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.String("port", a.port),
		slog.String("username", a.username),
		slog.String("db", a.db),
	)

	err := doWithAttempts(func() error {
		err := a.Pool.Ping(ctx)
		if err != nil {
			return err
		}
		return nil
	}, ctx, log, maxAttempts, delay)
	if err != nil {
		return fmt.Errorf("all attempts expired, failed to connect to postgresql: %w", err)
	}

	log.Info("connection to postgresql database is established")

	return nil
}

func (a *App) Stop() {
	const op = "pgapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("aborting postgresql connection...",
			slog.String("host", a.host),
			slog.String("port", a.port),
		)

	a.Pool.Close()
}

func doWithAttempts(f func() error, ctx context.Context, log *slog.Logger, attempts int, delay time.Duration) error {
	var err error
	for i := range attempts {
		err = f()
		if err == nil {
			return nil
		}

		if i == attempts-1 {
			break
		}

		log.Error("failed to ping to postgresql, retrying...", sl.Err(err))
		t := time.NewTimer(delay)

		select {
		case <-t.C:
			continue
		case <-ctx.Done():
			log.Error("context is cancelled, stopping retries")
			return fmt.Errorf("context is cancelled, stopping")
		}
	}

	log.Error("all attempts expired, failed to connect to postgresql", sl.Err(err))
	return fmt.Errorf("all attempts expired, failed to connect to postgresql: %w", err)
}
