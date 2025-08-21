package logger

import (
	"go.uber.org/zap"
)

// Log — глобальный логгер приложения.
// По умолчанию установлен в "no-op" режим до инициализации.
var Log = zap.NewNop()

// Initialize настраивает глобальный логгер на основе указанного уровня логирования.
//
// Параметры:
//   - level: строковое представление уровня логирования (например, "debug", "info").
//
// Возвращает:
//   - error: nil, если инициализация прошла успешно, иначе — ошибку.
func Initialize(level string) error {
	loggingLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = loggingLevel
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
