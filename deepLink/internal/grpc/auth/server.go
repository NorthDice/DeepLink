package auth

import (
	"context"
	"errors"
	"github.com/NorthDice/DeepLink/internal/services/auth"
	authv1 "github.com/NorthDice/DeepLink/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appId int32,
	) (token string, err error)
	RegisterNewUser(ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (isAdmin bool, err error)
}
type serverAPI struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(grpc *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(grpc, &serverAPI{auth: auth})
}

const (
	emptyValue  = 0
	emptyString = ""
)

func (s *serverAPI) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {

	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int32(req.GetAppId()))
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
func (s *serverAPI) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())

	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *authv1.IsAdminRequest,
) (*authv1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(req *authv1.LoginRequest) error {
	if req.Email == emptyString {
		return status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.Password == emptyString {
		return status.Error(codes.InvalidArgument, "password is empty")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app id is empty")
	}

	return nil
}

func validateRegister(req *authv1.RegisterRequest) error {
	if req.Email == emptyString {
		return status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.Password == emptyString {
		return status.Error(codes.InvalidArgument, "password is empty")
	}

	return nil
}
func validateIsAdmin(req *authv1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user id is empty")
	}
	return nil
}
