// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"os"
)

// Environment names.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Log formats.
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// Config holds all runtime configuration.
type Config struct {
	AppEnv         string
	HTTPAddr       string
	LogLevel       slog.Level
	LogFormat      string
	DatabaseURL    string
	MigrateOnStart bool
}

// Load reads configuration from the environment, applying defaults and validating values.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:    getenv("APP_ENV", EnvDevelopment),
		HTTPAddr:  getenv("HTTP_ADDR", ":8080"),
		LogFormat: getenv("LOG_FORMAT", LogFormatJSON),
	}

	switch cfg.AppEnv {
	case EnvDevelopment, EnvProduction:
	default:
		return Config{}, fmt.Errorf("config: invalid APP_ENV %q (want %q or %q)", cfg.AppEnv, EnvDevelopment, EnvProduction)
	}

	level, err := parseLogLevel(getenv("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	switch cfg.LogFormat {
	case LogFormatJSON, LogFormatText:
	default:
		return Config{}, fmt.Errorf("config: invalid LOG_FORMAT %q (want %q or %q)", cfg.LogFormat, LogFormatJSON, LogFormatText)
	}

	databaseURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok || databaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}
	cfg.DatabaseURL = databaseURL

	migrateOnStart, err := parseBool(getenv("MIGRATE_ON_START", defaultMigrateOnStart(cfg.AppEnv)))
	if err != nil {
		return Config{}, err
	}
	cfg.MigrateOnStart = migrateOnStart

	return cfg, nil
}

func defaultMigrateOnStart(appEnv string) string {
	if appEnv == EnvProduction {
		return "false"
	}
	return "true"
}

func parseBool(s string) (bool, error) {
	switch s {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("config: invalid MIGRATE_ON_START %q (want %q or %q)", s, "true", "false")
	}
}

func parseLogLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("config: invalid LOG_LEVEL %q (want debug, info, warn or error)", s)
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
