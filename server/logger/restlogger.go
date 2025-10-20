package logger

import (
	"bctbackend/logging"
)

// RestLogger is used by all REST endpoint to perform logging
type RestLogger interface {
	InvalidInput(message string)
	InvalidRequest(message string)
	InternalError(message string)
	Bug(message string)

	AddInformation(key string, value any) RestLogger
}

type LoggerWrapper struct {
	Logger logging.Logger
}

func NewLoggerWrapper(logger logging.Logger) RestLogger {
	return &LoggerWrapper{Logger: logger}
}

func (l *LoggerWrapper) InvalidInput(message string) {
	l.Logger.Warn(message)
}

func (l *LoggerWrapper) InvalidRequest(message string) {
	l.Logger.Warn(message)
}

func (l *LoggerWrapper) InternalError(message string) {
	l.Logger.Error(message)
}

func (l *LoggerWrapper) Bug(message string) {
	l.Logger.Error(message)
}

func (l *LoggerWrapper) AddInformation(key string, value any) RestLogger {
	l.Logger = l.Logger.With(key, value)
	return l
}
