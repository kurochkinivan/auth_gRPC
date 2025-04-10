package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kurochkinivan/auth/internal/entity"
	"github.com/kurochkinivan/auth/internal/lib/jwt"
	"github.com/kurochkinivan/auth/internal/lib/sl"
	"github.com/kurochkinivan/auth/internal/usecase/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user exists")
	ErrUserNotFound       = errors.New("user not found")
)

type Auth struct {
	log          *slog.Logger
	secret       string
	userSaver    UserSaver
	userProvider UserProvider
	tokentTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (userID uuid.UUID, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (*entity.User, error)
}

// New returns new instance of Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, secret string, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		secret:       secret,
		userSaver:    userSaver,
		userProvider: userProvider,
		tokentTTL:    tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system
//
// If exists, but password is incorrect, returns error.
// If user doesn't exist, returns error
func (a *Auth) Login(ctx context.Context, email, password string) (token string, err error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in successfully")

	token, err = jwt.NewToken(user, a.secret, a.tokentTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user id.
// If a user with given email already exists, returns error
func (a *Auth) RegisterNewUser(ctx context.Context, email, password string) (userID uuid.UUID, err error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	userID, err = a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			log.Warn("user exists", sl.Err(err))

			return uuid.Nil, ErrUserExists
		}

		log.Error("failed to save user", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return userID, nil
}
