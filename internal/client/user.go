package client

import (
	"context"

	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
)

var _ UserAPI = (*userClient)(nil)

type userClient struct {
	pbClient pb.GophKeeperClient
}

func (u *userClient) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	return u.pbClient.RegisterUser(ctx, req)
}

func (u *userClient) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return u.pbClient.Login(ctx, req)
}
