package service

import (
	"context"
	"github.com/faust8888/GophKeeper/internal/client"
	"github.com/faust8888/GophKeeper/internal/client/config"
	"github.com/faust8888/GophKeeper/internal/client/model"
	"github.com/faust8888/GophKeeper/internal/client/repository"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/faust8888/GophKeeper/internal/pkg/security"
)

// ==== Public Interfaces ====

// UserService defines operations related to user management.
type UserService interface {
	Register(ctx context.Context, req *pb.RegisterUserRequest) error
	Login(ctx context.Context, req *pb.LoginRequest) error
}

// SecretService defines operations related to secrets management.
type SecretService interface {
	AddSecret(ctx context.Context, secretType, description string, data []byte) error
	GetSecret(ctx context.Context, secretID string) (*model.Secret, error)
}

// ==== Facade ====

// ClientService is the main facade that provides access to user and secret operations.
// It implements both UserService and SecretService interfaces.
type ClientService struct {
	UserService   // embedded interface
	SecretService // embedded interface

	client  client.Client
	repo    repository.Repository
	session *model.Session
	crypto  *security.CryptoService
	cfg     *config.Config
}

// New creates a new ClientService with initialized sub-services.
func New(cfg *config.Config, repo repository.Repository, client client.Client) *ClientService {
	session := repo.GetSession()

	userSrv := newUserService(repo, client)
	secretSrv := newSecretService(cfg, repo, client)
	crypto := security.New(cfg.SecretKey, session.Salt)

	return &ClientService{
		UserService:   &userSrv,
		SecretService: &secretSrv,
		client:        client,
		repo:          repo,
		session:       session,
		crypto:        crypto,
		cfg:           cfg,
	}
}
