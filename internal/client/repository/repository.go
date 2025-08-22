package repository

import (
	"github.com/faust8888/GophKeeper/internal/client/model"
	"time"
)

type Repository interface {
	SaveSession(session *model.Session, expires time.Time) error
	GetSession() *model.Session
	SaveSecret(secret *model.Secret) error
	GetSecret(id string) (*model.Secret, error)
	GetSecretsByType(secretType string) ([]*model.Secret, error)
	GetAllSecrets() ([]*model.Secret, error)
	Close() error
}
