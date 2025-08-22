package model

import "time"

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
