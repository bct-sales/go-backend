package logging

import "log/slog"

type Logger interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
	With(key string, value any) Logger
}

type SloggerWrapper struct {
	wrapped *slog.Logger
}

func NewSloggerWrapper(logger *slog.Logger) Logger {
	return &SloggerWrapper{
		wrapped: logger,
	}
}

func (w *SloggerWrapper) Debug(message string) {
	w.wrapped.Debug(message)
}

func (w *SloggerWrapper) Info(message string) {
	w.wrapped.Info(message)
}

func (w *SloggerWrapper) Warn(message string) {
	w.wrapped.Warn(message)
}

func (w *SloggerWrapper) Error(message string) {
	w.wrapped.Error(message)
}

func (w *SloggerWrapper) With(key string, value any) Logger {
	return &SloggerWrapper{w.wrapped.With(key, value)}
}
