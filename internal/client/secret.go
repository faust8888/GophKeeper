package client

import (
	"context"

	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
)

var _ SecretAPI = (*secretClient)(nil)

type secretClient struct {
	pbClient pb.GophKeeperClient
}

func (s *secretClient) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	return s.pbClient.Sync(ctx, req)
}

func (s *secretClient) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	return s.pbClient.GetSecret(ctx, req)
}
