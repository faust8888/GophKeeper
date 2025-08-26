package sqllite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/faust8888/GophKeeper/internal/client/model"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(storageName string) (*Repository, error) {
	db, err := sql.Open("sqlite3", storageName)
	if err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS secrets (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			metadata TEXT NOT NULL,
			data BLOB NOT NULL,
			version INTEGER NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			salt TEXT NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_secrets_type ON secrets(type);
		CREATE INDEX IF NOT EXISTS idx_secrets_updated ON secrets(updated_at);
	`)
	return err
}

func (s *Repository) SaveSession(session *model.Session, expires time.Time) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO sessions (token, user_id, expires_at, salt) VALUES (?, ?, ?, ?)",
		session.Token, session.UserID, expires, session.Salt,
	)
	return err
}

func (s *Repository) GetSession() *model.Session {
	row := s.db.QueryRow("SELECT token, user_id, expires_at FROM sessions LIMIT 1")
	session := &model.Session{}
	err := row.Scan(&session.Token, &session.UserID, &session.Expires)
	if errors.Is(err, sql.ErrNoRows) {
		return &model.Session{}
	}
	return session
}

func (s *Repository) SaveSecret(secret *model.Secret) error {
	meta, err := json.Marshal(secret.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT OR REPLACE INTO secrets (id, type, metadata, data, version, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		secret.ID,
		secret.Type,
		string(meta),
		secret.Data,
		secret.Version,
		secret.UpdatedAt,
	)
	return err
}

func (s *Repository) GetSecret(id string) (*model.Secret, error) {
	row := s.db.QueryRow(
		"SELECT id, type, metadata, data, version, updated_at FROM secrets WHERE id = ?",
		id,
	)

	var sec model.Secret
	var metaStr string
	err := row.Scan(
		&sec.ID,
		&sec.Type,
		&metaStr,
		&sec.Data,
		&sec.Version,
		&sec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(metaStr), &sec.Metadata); err != nil {
		return nil, err
	}

	return &sec, nil
}

func (s *Repository) GetSecretsByType(secretType string) ([]*model.Secret, error) {
	rows, err := s.db.Query(
		"SELECT id, type, metadata, data, version, updated_at FROM secrets WHERE type = ?",
		secretType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*model.Secret
	for rows.Next() {
		var (
			sec     model.Secret
			metaStr string
		)
		if err = rows.Scan(
			&sec.ID,
			&sec.Type,
			&metaStr,
			&sec.Data,
			&sec.Version,
			&sec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(metaStr), &sec.Metadata); err != nil {
			return nil, err
		}
		sCopy := sec
		secrets = append(secrets, &sCopy)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}

func (s *Repository) GetAllSecrets() ([]*model.Secret, error) {
	rows, err := s.db.Query(
		"SELECT id, type, metadata, data, version, updated_at FROM secrets",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*model.Secret
	for rows.Next() {
		var (
			sec     model.Secret
			metaStr string
		)
		if err = rows.Scan(
			&sec.ID,
			&sec.Type,
			&metaStr,
			&sec.Data,
			&sec.Version,
			&sec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(metaStr), &sec.Metadata); err != nil {
			return nil, err
		}
		sCopy := sec
		secrets = append(secrets, &sCopy)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}

// Close - Закрываем хранилище.
func (s *Repository) Close() error {
	return s.db.Close()
}
