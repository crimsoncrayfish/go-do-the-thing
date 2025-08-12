package logger

import (
	"context"
	"log/slog"
	"os"
)

const (
	DEV  = "development"
	PROD = "production"
)

func SetupLogger() {
	env := os.Getenv("ENV")

	if env == PROD {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		slog.SetDefault(slog.New(handler))
	} else {
		handler := newColoredHandler(&slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(handler))
	}
}

func TODO(msg string, args ...any) {
	slog.Log(context.Background(), LevelTODO, msg, args...)
}

func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	panic(1)
}

const (
	LevelTODO  = slog.Level(-8) // Lower than Debug (-4)
	LevelFATAL = slog.Level(12) // Higher than Error (8)
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(source string) *Logger {
	return &Logger{
		logger: slog.Default().With("source", source),
	}
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) Todo(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelTODO, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelFATAL, msg, args...)
	panic(1)
}
