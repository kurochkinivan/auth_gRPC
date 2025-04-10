package pgclient

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(ctx context.Context, username, password, host, port, db string) (*pgxpool.Pool, error) {
	const op = "pgclient.NewClient"

	connString := &url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(username, password),
		Host:   net.JoinHostPort(host, port),
		Path:   db,
	}

	config, err := pgxpool.ParseConfig(connString.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pool, nil
}
