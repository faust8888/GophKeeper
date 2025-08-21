package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/faust8888/GophKeeper/internal/client"
	c "github.com/faust8888/GophKeeper/internal/client/config"
	"github.com/faust8888/GophKeeper/internal/client/repository"
	"github.com/faust8888/GophKeeper/internal/pkg/common"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/faust8888/GophKeeper/internal/pkg/security"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

var ErrRequiredArgumentIsMissing = errors.New("one of required arguments is missing")

type Service struct {
	client  client.Client
	repo    repository.Repository
	session *repository.Session
	crypto  *security.CryptoService
	config  *c.Config
}

func New(config *c.Config, repo repository.Repository, client client.Client) (*Service, error) {
	session := repo.GetSession()
	return &Service{
		client:  client,
		repo:    repo,
		session: session,
		crypto:  security.New(config.SecretKey, session.Salt),
		config:  config,
	}, nil
}

func (s *Service) SignUpUser(ctx context.Context, req *pb.RegisterUserRequest) error {
	if req.Login == "" || len(req.Password) == 0 {
		return ErrRequiredArgumentIsMissing
	}

	_, err := s.client.RegisterUser(ctx, req)
	if err != nil {
		return fmt.Errorf("server failed to sign up req: %w", err)
	}

	return nil
}

func (s *Service) SignInUser(ctx context.Context, req *pb.LoginRequest) error {
	if req.Login == "" || req.Password == "" {
		return ErrRequiredArgumentIsMissing
	}

	res, err := s.client.Login(ctx, req)
	if err != nil {
		return fmt.Errorf("server failed to sign in req: %w", err)
	}

	session := repository.Session{
		Token:  res.Token,
		UserID: res.UserId,
		Salt:   security.GenerateSalt(),
	}

	if err := s.repo.SaveSession(session, time.Now().Add(12*time.Hour)); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (s *Service) AddSecret(ctx context.Context, secretType, description string, data []byte) error {
	encryptedData, err := s.crypto.Encrypt(data)
	if err != nil {
		return err
	}

	_, err = s.crypto.Decrypt(encryptedData)
	if err != nil {
		fmt.Println("could not decrypt data")
	}

	metadata := map[string]string{
		"description": description,
		"created_at":  time.Now().Format(time.RFC3339),
	}

	metadata["type"] = secretType

	localSecret := &repository.Secret{
		ID:        uuid.New().String(),
		Type:      secretType,
		Metadata:  metadata,
		Data:      encryptedData,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	if err = s.repo.SaveSecret(localSecret); err != nil {
		return err
	}

	localSecrets, err := s.repo.GetAllSecrets()
	if err != nil {
		return err
	}

	var secrets []*pb.Secret
	for _, s := range localSecrets {
		secrets = append(secrets, &pb.Secret{
			Id:        s.ID,
			Type:      common.SecretTypeFromString(s.Type),
			Metadata:  s.Metadata,
			Data:      s.Data,
			Version:   s.Version,
			UpdatedAt: timestamppb.New(s.UpdatedAt),
		})
	}

	resp, err := s.client.Sync(ctx, &pb.SyncRequest{
		Token:        s.session.Token,
		LocalSecrets: secrets,
	})
	if err != nil {
		return err
	}

	for _, responseSecret := range resp.ServerSecrets {
		if err = s.repo.SaveSecret(&repository.Secret{
			ID:        responseSecret.Id,
			Type:      common.SecretTypeToString(responseSecret.Type),
			Metadata:  responseSecret.Metadata,
			Data:      responseSecret.Data,
			Version:   responseSecret.Version,
			UpdatedAt: responseSecret.UpdatedAt.AsTime(),
		}); err != nil {
			log.Printf("Failed to save secret %s: %v", responseSecret.Id, err)
		}
	}
	return nil
}

func (s *Service) GetSecret(ctx context.Context, secretID string) (*repository.Secret, error) {
	localSecret, err := s.repo.GetSecret(secretID)
	if err == nil {
		decrypted, err := s.crypto.Decrypt(localSecret.Data)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
		localSecret.Data = decrypted
		return localSecret, nil
	}

	resp, err := s.client.GetSecret(ctx, &pb.GetSecretRequest{
		Token:    s.repo.GetSession().Token,
		SecretId: secretID,
	})
	if err != nil {
		return nil, fmt.Errorf("server error: %w", err)
	}

	decrypted, err := s.crypto.Decrypt(resp.Secret.Data)
	if err != nil {
		return nil, fmt.Errorf("decryption failed:: %w", err)
	}

	secret := &repository.Secret{
		ID:        resp.Secret.Id,
		Type:      common.SecretTypeToString(resp.Secret.Type),
		Metadata:  resp.Secret.Metadata,
		Data:      decrypted,
		Version:   resp.Secret.Version,
		UpdatedAt: resp.Secret.UpdatedAt.AsTime(),
	}

	if err = s.repo.SaveSecret(secret); err != nil {
		log.Printf("Warning: failed to cache secret locally: %v", err)
	}

	return secret, nil
}
