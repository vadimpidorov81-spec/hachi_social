package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Environment     string
	Version         string
	HTTPAddr        string
	DatabaseURL     string
	DevelopmentUser string
	LogLevel        string
	LogDir          string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	shutdownTimeout, err := time.ParseDuration(value("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
	}

	cfg := Config{
		Environment:     value("APP_ENV", "development"),
		Version:         value("APP_VERSION", "dev"),
		HTTPAddr:        value("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		DevelopmentUser: os.Getenv("APP_DEV_USER_ID"),
		LogLevel:        value("LOG_LEVEL", "info"),
		LogDir:          value("LOG_DIR", "out/logs"),
		ShutdownTimeout: shutdownTimeout,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.Environment != "development" && cfg.DevelopmentUser != "" {
		return Config{}, errors.New("APP_DEV_USER_ID is allowed only in development")
	}

	return cfg, nil
}

func value(key, fallback string) string {
	if current := os.Getenv(key); current != "" {
		return current
	}
	return fallback
}
