package repository

import (
	"time"
)

type Repository interface {
	SaveSession(session Session, expires time.Time) error
	GetSession() *Session
	SaveSecret(secret *Secret) error
	GetSecret(id string) (*Secret, error)
	GetSecretsByType(secretType string) ([]*Secret, error)
	GetAllSecrets() ([]*Secret, error)
	Close() error
}

type Secret struct {
	ID        string
	Type      string
	Metadata  map[string]string
	Data      []byte
	Version   int32
	UpdatedAt time.Time
}

type Session struct {
	Token   string
	UserID  string
	Expires time.Time
	Salt    string
}
