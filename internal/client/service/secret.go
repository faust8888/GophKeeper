package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/faust8888/GophKeeper/internal/client"
	"github.com/faust8888/GophKeeper/internal/client/config"
	"github.com/faust8888/GophKeeper/internal/client/model"
	"github.com/faust8888/GophKeeper/internal/client/repository"
	"github.com/faust8888/GophKeeper/internal/pkg/common"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/faust8888/GophKeeper/internal/pkg/security"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// private implementation of SecretService
type secretService struct {
	client  client.Client
	repo    repository.Repository
	session *model.Session
	crypto  *security.CryptoService
}

func newSecretService(cfg *config.Config, repo repository.Repository, client client.Client) secretService {
	session := repo.GetSession()
	crypto := security.New(cfg.SecretKey, session.Salt)
	return secretService{client, repo, session, crypto}
}

func (s *secretService) AddSecret(ctx context.Context, secretType, description string, data []byte) error {
	encryptedData, err := s.crypto.Encrypt(data)
	if err != nil {
		return err
	}

	metadata := map[string]string{
		"description": description,
		"created_at":  time.Now().Format(time.RFC3339),
		"type":        secretType,
	}

	localSecret := &model.Secret{
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
		if err = s.repo.SaveSecret(&model.Secret{
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

func (s *secretService) GetSecret(ctx context.Context, secretID string) (*model.Secret, error) {
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
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	secret := &model.Secret{
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
