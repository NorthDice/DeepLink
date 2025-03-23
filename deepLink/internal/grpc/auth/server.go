package auth

import (
	"context"
	authv1 "github.com/NorthDice/DeepLink/protos/gen/go/auth"
	"google.golang.org/grpc"
)

type serverAPI struct {
	authv1.UnimplementedAuthServer
}

func Register(grpc *grpc.Server) {
	authv1.RegisterAuthServer(grpc, &serverAPI{})
}
func (s *serverAPI) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	panic("implement me")
}
func (s *serverAPI) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	panic("implement me")
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *authv1.IsAdminRequest,
) (*authv1.IsAdminResponse, error) {
	panic("implement me")
}
