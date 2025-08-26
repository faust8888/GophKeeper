package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/faust8888/GophKeeper/internal/server/config"
	"github.com/faust8888/GophKeeper/internal/server/model"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// ErrNoRowWasAffected — ошибка, возникающая при попытке дублирования записи по full_url.
var ErrNoRowWasAffected = errors.New("no row was affected")

var ErrNotFound = errors.New("not found")

// Repository — реализация repository.Repository на основе PostgreSQL.
// Используется для хранения, поиска и удаления коротких ссылок в БД.
type Repository struct {
	db *sql.DB // Подключение к базе данных
}

func (repo *Repository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	const query = `SELECT id, login, password_hash, salt FROM "user" WHERE login = $1`

	var user model.User
	err := repo.db.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.Salt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "Not found client")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (repo *Repository) CreateUser(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO "user" (id, login, password_hash, salt)
		VALUES ($1, $2, $3, $4)
	`

	user.ID = uuid.New()

	res, err := repo.db.ExecContext(ctx, query, user.ID, user.Login, user.PasswordHash, user.Salt)
	if err != nil {
		return fmt.Errorf("repository.postgres.CreateUser: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrNoRowWasAffected
	}
	return nil
}

func (repo *Repository) GetSecretById(ctx context.Context, userID uuid.UUID, secretID string) (*model.Secret, error) {
	const query = `
		SELECT id, user_id, type, metadata, data, version, created_at, updated_at 
		FROM secret 
		WHERE id = $1 AND user_id = $2
	`

	var (
		sec      model.Secret
		metaJSON []byte
	)

	// Проверяем что secretID - валидный UUID
	if _, err := uuid.Parse(secretID); err != nil {
		return nil, fmt.Errorf("invalid secret ID format: %w", err)
	}

	row := repo.db.QueryRowContext(ctx, query, secretID, userID)

	err := row.Scan(
		&sec.ID,
		&sec.UserID,
		&sec.Type,
		&metaJSON,
		&sec.Data,
		&sec.Version,
		&sec.CreatedAt,
		&sec.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query secret: %w", err)
	}

	// Декодируем метаданные из JSON
	if err := json.Unmarshal(metaJSON, &sec.Metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	// Дополнительная проверка принадлежности пользователю
	if sec.UserID != userID {
		return nil, fmt.Errorf("err access denaid: %w", err)
	}

	return &sec, nil
}

func (repo *Repository) GetSecrets(ctx context.Context, userID uuid.UUID) ([]*model.Secret, error) {
	const query = `
		SELECT id, user_id, type, metadata, data, version, created_at, updated_at
		FROM secret
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	var secrets []*model.Secret
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sec model.Secret
		var meta []byte

		if err = rows.Scan(
			&sec.ID,
			&sec.UserID,
			&sec.Type,
			&meta,
			&sec.Data,
			&sec.Version,
			&sec.CreatedAt,
			&sec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}

		if err = json.Unmarshal(meta, &sec.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		secrets = append(secrets, &sec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return secrets, nil
}

func (repo *Repository) UpsertSecret(ctx context.Context, secret *model.Secret) error {
	const query = `
		INSERT INTO secret (id, user_id, type, metadata, data, version, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			metadata = EXCLUDED.metadata,
			data = EXCLUDED.data,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
		WHERE secret.version < EXCLUDED.version
	`

	// Подготовка метаданных
	meta, err := json.Marshal(secret.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Увеличиваем версию при обновлении
	secret.Version++
	secret.UpdatedAt = time.Now()

	_, err = repo.db.ExecContext(ctx, query, secret.ID,
		secret.UserID,
		secret.Type,
		meta,
		secret.Data,
		secret.Version,
		secret.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert secret: %w", err)
	}

	return nil
}

// NewPostgresRepository создаёт новый экземпляр Repository, подключаясь к PostgreSQL.
//
// Паникует, если не может установить соединение.
//
// Параметр:
//   - cfg: конфигурация приложения.
//
// Возвращает:
//   - *Repository: готовый к использованию объект репозитория.
func NewPostgresRepository(cfg *config.Config) (*Repository, error) {
	db, err := sql.Open("pgx", cfg.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("postgres.NewPostgresRepository: %w", err)
	}
	return &Repository{
		db: db,
	}, nil
}
