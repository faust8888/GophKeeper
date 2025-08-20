package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/faust8888/GophKeeper/internal/pkg/logger"
	"github.com/faust8888/GophKeeper/internal/server/config"
	"github.com/faust8888/GophKeeper/internal/server/model"
	pb "github.com/faust8888/GophKeeper/internal/server/proto/gophkeeperpb"
	"github.com/faust8888/GophKeeper/internal/server/repository"
	"github.com/faust8888/GophKeeper/internal/server/security"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type GophKeeperService struct {
	pb.UnimplementedGophKeeperServer
	repo repository.Repository
	cfg  *config.Config
}

func NewGophKeeperService(cfg *config.Config, repo repository.Repository) *GophKeeperService {
	return &GophKeeperService{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *GophKeeperService) Register(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	if _, err := s.repo.GetUserByLogin(ctx, req.Login); err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	hash := security.HashPassword(req.Password)
	user := model.User{
		Login:        req.Login,
		PasswordHash: hash,
	}
	if err := s.repo.CreateUser(ctx, &user); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.RegisterUserResponse{UserId: user.ID.String()}, nil
}

func (s *GophKeeperService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.repo.GetUserByLogin(ctx, req.Login)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if err = security.CompareHash(user.PasswordHash, req.Password); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	userID := user.ID.String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.cfg.AuthKey))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: tokenString, UserId: userID}, nil
}

func (s *GophKeeperService) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	userID, err := security.Authenticate(req.Token, s.cfg.AuthKey)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	serverSecrets, err := s.repo.GetSecrets(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get secrets")
	}

	serverVersions := make(map[string]int32)
	for _, sec := range serverSecrets {
		serverVersions[sec.ID.String()] = sec.Version
	}

	for _, local := range req.LocalSecrets {
		localID, err := uuid.Parse(local.Id)
		if err != nil {
			continue
		}

		if serverVer, exists := serverVersions[local.Id]; exists && local.Version <= serverVer {
			continue
		}

		secret := model.Secret{
			ID:       localID,
			UserID:   uid,
			Type:     SecretTypeToString(local.Type),
			Metadata: local.Metadata,
			Data:     local.Data,
			Version:  local.Version,
		}
		if err := s.repo.UpsertSecret(ctx, &secret); err != nil {
			logger.Log.Error("Failed to upsert secret", zap.Error(err))
		}
	}

	updatedSecrets, err := s.repo.GetSecrets(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get updated secrets")
	}

	pbSecrets := make([]*pb.Secret, 0, len(updatedSecrets))
	for _, sec := range updatedSecrets {
		pbSecrets = append(pbSecrets, &pb.Secret{
			Id:        sec.ID.String(),
			Type:      SecretTypeFromString(sec.Type),
			Metadata:  sec.Metadata,
			Data:      sec.Data,
			Version:   sec.Version,
			UpdatedAt: timestamppb.New(sec.UpdatedAt),
		})
	}

	return &pb.SyncResponse{ServerSecrets: pbSecrets}, nil
}

func (s *GophKeeperService) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	userID, err := security.Authenticate(req.Token, s.cfg.AuthKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}
	secret, err := s.repo.GetSecretById(ctx, uid, req.SecretId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	return &pb.GetSecretResponse{
		Secret: &pb.Secret{
			Id:        secret.ID.String(),
			Type:      SecretTypeFromString(secret.Type),
			Metadata:  secret.Metadata,
			Data:      secret.Data,
			Version:   secret.Version,
			UpdatedAt: timestamppb.New(secret.UpdatedAt),
		},
	}, nil
}

func (s *GophKeeperService) mustEmbedUnimplementedGophKeeperServer() {}

func SecretTypeFromString(s string) pb.SecretType {
	switch s {
	case "EMPTY":
		return pb.SecretType_EMPTY
	case "UNSPECIFIED":
		return pb.SecretType_UNSPECIFIED
	case "CREDENTIALS":
		return pb.SecretType_CREDENTIALS
	case "TEXT":
		return pb.SecretType_TEXT
	case "BINARY":
		return pb.SecretType_BINARY
	case "CARD":
		return pb.SecretType_CARD
	default:
		return pb.SecretType_UNSPECIFIED
	}
}

func SecretTypeToString(secretType pb.SecretType) string {
	switch secretType {
	case pb.SecretType_EMPTY:
		return "EMPTY"
	case pb.SecretType_UNSPECIFIED:
		return "UNSPECIFIED"
	case pb.SecretType_CREDENTIALS:
		return "CREDENTIALS"
	case pb.SecretType_TEXT:
		return "TEXT"
	case pb.SecretType_BINARY:
		return "BINARY"
	case pb.SecretType_CARD:
		return "CARD"
	default:
		return "UNSPECIFIED"
	}
}
