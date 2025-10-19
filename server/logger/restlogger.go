package logger

import (
	"log/slog"
)

// RestLogger is used by all REST endpoint to perform logging
type RestLogger interface {
	InvalidInput(message string, args ...any)
	InvalidRequest(message string, args ...any)
	InternalError(message string, args ...any)
	Bug(message string, args ...any)

	AddInformation(key string, value any)
}

type LoggerWrapper struct {
	Logger *slog.Logger
}

func NewLoggerWrapper(logger *slog.Logger) *LoggerWrapper {
	return &LoggerWrapper{Logger: logger}
}

func (l *LoggerWrapper) InvalidInput(message string, args ...any) {
	l.Logger.Warn(message, args...)
}

func (l *LoggerWrapper) InvalidRequest(message string, args ...any) {
	l.Logger.Warn(message, args...)
}

func (l *LoggerWrapper) InternalError(message string, args ...any) {
	l.Logger.Error(message, args...)
}

func (l *LoggerWrapper) Bug(message string, args ...any) {
	l.Logger.Error(message, args...)
}

func (l *LoggerWrapper) AddInformation(key string, value any) {
	l.Logger = l.Logger.With(key, value)
}
