package auth

import authv1 "github.com/NorthDice/DeepLink/protos/gen/go/auth"

type serverAPI struct {
	authv1.UnimplementedAuthServer
}
