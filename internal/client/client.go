package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
)

// ==== Sub-interfaces ====

// UserAPI defines user-related RPCs.
type UserAPI interface {
	RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error)
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
}

// SecretAPI defines secret-related RPCs.
type SecretAPI interface {
	Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error)
	GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error)
}

// ==== Main Client Interface ====

// Client combines both UserAPI and SecretAPI.
type Client interface {
	UserAPI
	SecretAPI
	Close() error
}

var _ Client = (*GRPCClient)(nil)

// ==== Implementation ====

// GRPCClient is a concrete implementation of Client, composed of sub-clients.
type GRPCClient struct {
	UserAPI
	SecretAPI
	conn *grpc.ClientConn
}

// NewGRPCClient connects to the gRPC server and builds a GRPCClient.
func NewGRPCClient(serverAddress string) (*GRPCClient, error) {
	conn, err := grpc.Dial(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	pbClient := pb.NewGophKeeperClient(conn)

	return &GRPCClient{
		UserAPI:   &userClient{pbClient: pbClient},
		SecretAPI: &secretClient{pbClient: pbClient},
		conn:      conn,
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
