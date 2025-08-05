package logger

import "log/slog"

type Logger interface {
	InvalidInput(message string, args ...any)
	InvalidRequest(message string, args ...any)
	DatabaseError(message string, args ...any)
	Bug(message string, args ...any)
}

type LoggerWrapper struct {
	logger *slog.Logger
}

func NewLoggerWrapper(logger *slog.Logger) *LoggerWrapper {
	return &LoggerWrapper{logger: logger}
}

func (l *LoggerWrapper) InvalidInput(message string, args ...any) {
	l.logger.Warn(message, args...)
}

func (l *LoggerWrapper) InvalidRequest(message string, args ...any) {
	l.logger.Warn(message, args...)
}

func (l *LoggerWrapper) DatabaseError(message string, args ...any) {
	l.logger.Error(message, args...)
}

func (l *LoggerWrapper) Bug(message string, args ...any) {
	l.logger.Error(message, args...)
}
