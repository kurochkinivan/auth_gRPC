package pg

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/auth/internal/entity"
	"github.com/kurochkinivan/auth/internal/usecase/repository"
	"github.com/kurochkinivan/auth/pkg/pgerr"
)

type Repository struct {
	pool *pgxpool.Pool
	qb   sq.StatementBuilderType
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
		qb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// SaveUser saves user in the database.
//
// If user with given email already exists, returns error repository.ErrUserExists.
// If an error occurs during query execution, returns an error.
func (r *Repository) SaveUser(ctx context.Context, email string, passHash []byte) (userID uuid.UUID, err error) {
	const op = "repository.pg.SaveUser"

	sql, args, err := r.qb.
		Insert(TableUsers).
		Columns(
			"email",
			"password",
		).
		Values(
			email,
			passHash,
		).
		Suffix("ON CONFLICT DO NOTHING").
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, pgerr.ErrCreateQuery(op, err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, repository.ErrUserExists
		}
		return uuid.Nil, pgerr.ErrScan(op, err)
	}

	return userID, nil
}

// User returns a user by email from the database.
//
// If the user is not found, it returns repository.ErrUserNotFound.
// If an error occurs during query execution, it returns an error.
func (r *Repository) User(ctx context.Context, email string) (*entity.User, error) {
	const op = "repository.pg.User"

	sql, args, err := r.qb.
		Select(
			"id",
			"email",
			"password",
		).
		From(TableUsers).
		Where(
			sq.Eq{"email": email},
		).
		ToSql()
	if err != nil {
		return nil, pgerr.ErrCreateQuery(op, err)
	}

	user := new(entity.User)
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PassHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, pgerr.ErrScan(op, err)
	}

	return user, nil
}
