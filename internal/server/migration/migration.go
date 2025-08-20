package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/faust8888/GophKeeper/internal/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run применяет SQL-миграции к указанной PostgreSQL базе данных.
//
// Метод:
// - Устанавливает соединение с БД.
// - Создаёт драйвер миграции для PostgreSQL.
// - Инициализирует систему миграций, используя SQL-файлы из директории `internal/app/migration/sql`.
// - Применяет все новые миграции.
//
// Если миграции уже применены (ErrNoChange), это не считается ошибкой.
//
// Параметры:
//   - dataSourceName: строка подключения к PostgreSQL (DSN).
//
// Возвращает:
//   - error: nil, если миграции успешно применены или уже были применены ранее,
//     иначе — соответствующую ошибку.
func Run(dataSourceName string) error {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return fmt.Errorf("migration.run: oppening conection - %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migration.run: creating migration driver - %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/server/migration/sql",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migration.run: creating migration instance - %w", err)
	}

	if err = m.Up(); !errors.Is(err, migrate.ErrNoChange) && err != nil {
		return fmt.Errorf("migration.run: applying migrations - %w", err)
	}
	logger.Log.Info("migration.run: all sql scripts applied successfully")
	return nil
}
