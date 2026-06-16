package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Logger struct {
	*slog.Logger
	file *os.File
}

func New(directory, level, environment, version string) (*Logger, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	filename := filepath.Join(directory, time.Now().UTC().Format("2006-01-02")+".log")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	handler := slog.NewJSONHandler(io.MultiWriter(os.Stdout, file), &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	logger := slog.New(handler).With(
		slog.String("service", "hachisocial"),
		slog.String("environment", environment),
		slog.String("version", version),
	)

	return &Logger{Logger: logger, file: file}, nil
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
