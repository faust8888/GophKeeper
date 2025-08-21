package client

import (
	"context"
	"fmt"
	c "github.com/faust8888/GophKeeper/internal/client/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
)

type Client interface {
	RegisterUser(ctx context.Context, user *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error)
	Login(ctx context.Context, user *pb.LoginRequest) (*pb.LoginResponse, error)
	Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error)
	GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error)
}

var _ Client = (*GRPCClient)(nil)

type GRPCClient struct {
	token  string
	conn   *grpc.ClientConn
	config *c.Config
	client pb.GophKeeperClient
}

func (g *GRPCClient) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	resp, err := g.client.RegisterUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *GRPCClient) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp, err := g.client.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *GRPCClient) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	resp, err := g.client.Sync(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *GRPCClient) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	resp, err := g.client.GetSecret(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewGRPCClient(config *c.Config) (*GRPCClient, error) {
	client := &GRPCClient{
		config: config,
	}

	conn, err := grpc.NewClient(config.ServerGRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithUnaryInterceptor(client.authInterceptor),
	)
	if err != nil {
		fmt.Errorf("create grpc client error: %v", err)
	}

	client.conn = conn
	client.client = pb.NewGophKeeperClient(conn)

	return client, nil
}
