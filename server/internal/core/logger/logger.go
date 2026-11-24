package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type contextKey struct{}

type Config struct {
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func New(cfg Config) (*slog.Logger, io.Closer, error) {
	filePath := strings.TrimSpace(cfg.FilePath)
	if filePath == "" {
		return NewWithWriters(os.Stdout, nil, cfg), nopCloser{}, nil
	}

	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, fmt.Errorf("create log directory %q: %w", dir, err)
		}
	}

	rotator := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	return NewWithWriters(os.Stdout, rotator, cfg), rotator, nil
}

func NewWithWriters(stdout, file io.Writer, cfg Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	}

	if stdout == nil && file == nil {
		stdout = os.Stdout
	}

	handlers := make([]slog.Handler, 0, 2)
	if stdout != nil {
		handlers = append(handlers, slog.NewTextHandler(stdout, opts))
	}
	if file != nil {
		handlers = append(handlers, slog.NewJSONHandler(file, opts))
	}

	if len(handlers) == 1 {
		return slog.New(handlers[0])
	}

	return slog.New(slog.NewMultiHandler(handlers...))
}

func ToContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && log != nil {
		return log
	}

	return slog.Default()
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
