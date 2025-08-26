package repository

import (
	"context"
	"github.com/faust8888/GophKeeper/internal/server/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	GetSecretById(ctx context.Context, userID uuid.UUID, secretID string) (*model.Secret, error)
	GetSecrets(ctx context.Context, userID uuid.UUID) ([]*model.Secret, error)
	UpsertSecret(ctx context.Context, secret *model.Secret) error
}
