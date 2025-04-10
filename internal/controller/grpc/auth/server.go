package auth

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kurochkinivan/auth/internal/usecase/auth"
	authv1 "github.com/kurochkinivan/auth_proto/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email, password string) (token string, err error)
	RegisterNewUser(ctx context.Context, email, password string) (userID uuid.UUID, err error)
}

type serverAPI struct {
	authv1.UnimplementedAuthServer
	validate *validator.Validate
	auth     Auth
}

func Register(gRPC *grpc.Server, validate *validator.Validate, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{
		validate: validate,
		auth:     auth,
	})
}

func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if err := validateLogin(req, s.validate); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := validateRegister(req, s.validate); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RegisterResponse{
		UserId: userID.String(),
	}, nil
}

func validateLogin(req *authv1.LoginRequest, validate *validator.Validate) error {
	err := validate.Var(req.GetEmail(), "required,email")
	if err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrs {
				if e.Tag() == "required" {
					return status.Error(codes.InvalidArgument, "email is required")
				}
				if e.Tag() == "email" {
					return status.Error(codes.InvalidArgument, "invalid email format")
				}
			}
		}
	}

	err = validate.Var(req.GetPassword(), "required")
	if err != nil {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

func validateRegister(req *authv1.RegisterRequest, validate *validator.Validate) error {
	err := validate.Var(req.GetEmail(), "required,email")
	if err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrs {
				if e.Tag() == "required" {
					return status.Error(codes.InvalidArgument, "email is required")
				}
				if e.Tag() == "email" {
					return status.Error(codes.InvalidArgument, "invalid email format")
				}
			}
		}
	}

	err = validate.Var(req.GetPassword(), "required")
	if err != nil {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}
