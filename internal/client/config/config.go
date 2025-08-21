// Package config предоставляет управление конфигурацией для приложения.
// Он загружает настройки из командных флагов, переменных окружения и JSON-файла,
// обеспечивая правильный приоритет: флаги > переменные окружения > JSON-файл > значения по умолчанию.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/caarlos0/env/v6"
)

// Константы для флагов командной строки.
const (
	// ServerGRPCAddressFlag - флаг для адреса GRCP сервера (-a).
	ServerGRPCAddressFlag = "a"
	// LoggingLevelFlag - флаг для уровня логирования (-l).
	LoggingLevelFlag = "l"
	// DataSourceNameFlag - флаг для строки подключения к БД (-d).
	DataSourceNameFlag = "d"
	// SecretKeyNameFlag - флаг для ключа аутентификации (-k).
	SecretKeyNameFlag = "k"
	// EnableTLSOnServerFlag - флаг для включения HTTPS (-s).
	EnableTLSOnServerFlag = "s"
	// ConfigFileFlag - флаг для пути к файлу конфигурации (-c).
	ConfigFileFlag = "c"
	// ConfigFileFlagAlias - псевдоним флага для пути к файлу конфигурации (-config).
	ConfigFileFlagAlias = "config"
)

// Config хранит все настройки конфигурации приложения.
type Config struct {
	// ServerGRPCAddress - сетевой адрес и порт для запуска сервера (флаг -ag, env SERVER_GRCP_ADDRESS).
	ServerGRPCAddress string `env:"SERVER_GRCP_ADDRESS" json:"server_grpc_address"`
	// LoggingLevel - уровень логирования (флаг -l, env LOGGING_LEVEL).
	LoggingLevel string `env:"LOGGING_LEVEL"`
	// DataSourceName - строка подключения к базе данных PostgreSQL (флаг -d, env DATABASE_DSN).
	DataSourceName string `env:"DATABASE_DSN" json:"database_dsn"`
	// SecretKey - секретный ключ для подписи токенов аутентификации (флаг -k, env SECRET_KEY).
	SecretKey string `env:"SECRET_KEY"`
	// EnableHTTPS - флаг, включающий HTTPS на сервере (флаг -s, env ENABLE_HTTPS).
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
}

// JSONConfig - это вспомогательная структура для разбора конфигурации из JSON-файла.
// Использование указателей позволяет отличить отсутствующее в JSON поле от поля с нулевым значением
// (например, пустой строки или false).
type JSONConfig struct {
	ServerGRCPAddress *string `json:"server_grpc_address"`
	DataSourceName    *string `json:"database_dsn"`
	EnableHTTPS       *bool   `json:"enable_https"`
}

var (
	// cfg - глобальный синглтон-экземпляр конфигурации приложения.
	cfg *Config
	// once используется для гарантии того, что инициализация конфигурации произойдет только один раз.
	once sync.Once
)

// Create инициализирует и возвращает синглтон-объект конфигурации.
//
// Функция определяет настройки, считывая их из различных источников
// в следующем порядке приоритета (от высшего к низшему):
//  1. Флаги командной строки (например, -a, -b).
//  2. Переменные окружения (например, SERVER_ADDRESS, BASE_URL).
//  3. Файл конфигурации в формате JSON (путь к которому задается флагом -c/-config или переменной окружения CONFIG).
//  4. Значения по умолчанию, заданные в коде.
//
// Благодаря использованию sync.Once, логика инициализации выполняется только при первом вызове.
// Все последующие вызовы мгновенно возвращают уже настроенный экземпляр.
func Create() *Config {
	once.Do(func() {
		configFilePath := findConfigPathUsingFlags()

		if configFilePath == "" {
			configFilePath = os.Getenv("CONFIG")
		}

		cfg = defaultConfig()

		if configFilePath != "" {
			cfg.applyJSONConfig(configFilePath)
		}

		if err := env.Parse(cfg); err != nil {
			fmt.Errorf("failed to parse environment variables: %v", err)
		}

		defineGlobalFlags()

		flag.Parse()
	})

	return cfg
}

// defaultConfig создает новый экземпляр Config со значениями по умолчанию.
func defaultConfig() *Config {
	return &Config{
		ServerGRPCAddress: "localhost:8090",
		LoggingLevel:      "INFO",
		DataSourceName:    "client.db",
		SecretKey:         "dd109d0b86dc6a06584a835538768c6a2ceb588560755c7f7b90c0bf774237c8",
		EnableHTTPS:       false,
	}
}

// findConfigPathUsingFlags использует временный, изолированный FlagSet для поиска
// пути к файлу конфигурации в аргументах командной строки. Это позволяет найти путь
// до основного парсинга флагов и избежать ошибок о неопределенных флагах.
func findConfigPathUsingFlags() string {
	configFlagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	// Перенаправляем вывод ошибок этого временного FlagSet в "никуда",
	// чтобы он не засорял консоль при встрече с флагами, которые он не знает (-a, -b и т.д.).
	configFlagSet.SetOutput(io.Discard)

	var configPath string
	configFlagSet.StringVar(&configPath, ConfigFileFlag, "", "Path to JSON config file")
	configFlagSet.StringVar(&configPath, ConfigFileFlagAlias, "", "Path to JSON config file (alias)")

	// Парсим аргументы. Ошибки будут проигнорированы и не будут выведены в консоль.
	_ = configFlagSet.Parse(os.Args[1:])

	return configPath
}

// applyJSONConfig читает конфигурационный файл JSON по указанному пути
// и применяет его настройки к экземпляру Config.
// Метод перезаписывает поля только в том случае, если они присутствуют в JSON-файле.
func (c *Config) applyJSONConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var jsonCfg JSONConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		return
	}
	if jsonCfg.ServerGRCPAddress != nil {
		c.ServerGRPCAddress = *jsonCfg.ServerGRCPAddress
	}
	if jsonCfg.DataSourceName != nil {
		c.DataSourceName = *jsonCfg.DataSourceName
	}
	if jsonCfg.EnableHTTPS != nil {
		c.EnableHTTPS = *jsonCfg.EnableHTTPS
	}
}

// defineGlobalFlags определяет все флаги командной строки приложения в глобальном наборе flag.CommandLine.
// В качестве значений по умолчанию для флагов используются уже загруженные значения из cfg.
// Это обеспечивает правильный порядок приоритетов при вызове flag.Parse().
func defineGlobalFlags() {
	flag.StringVar(&cfg.ServerGRPCAddress, ServerGRPCAddressFlag, cfg.ServerGRPCAddress, "Address of the GRCP server (ex: localhost:8090)")
	flag.StringVar(&cfg.DataSourceName, DataSourceNameFlag, cfg.DataSourceName, "Data Source Name for PostgreSQL (ex: postgres://user:pass@host:port/db)")
	flag.BoolVar(&cfg.EnableHTTPS, EnableTLSOnServerFlag, cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.LoggingLevel, LoggingLevelFlag, cfg.LoggingLevel, "Level of logging to use")
	flag.StringVar(&cfg.SecretKey, SecretKeyNameFlag, cfg.SecretKey, "Secret Key")

	// Определяем флаг -c/-config здесь еще раз, чтобы он отображался в справке (-h).
	// Его значение нам уже не нужно, так как мы его получили ранее.
	var dummyConfigPath string
	flag.StringVar(&dummyConfigPath, ConfigFileFlag, "", "Path to JSON config file")
	flag.StringVar(&dummyConfigPath, ConfigFileFlagAlias, "", "Path to JSON config file (alias)")
}
