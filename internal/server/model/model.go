package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	Salt         string    `db:"salt"`
}

type Secret struct {
	ID        uuid.UUID         `db:"id"`
	UserID    uuid.UUID         `db:"user_id"`
	Type      string            `db:"type"`
	Metadata  map[string]string `db:"metadata"`
	Data      []byte            `db:"data"`
	Version   int32             `db:"version"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}
