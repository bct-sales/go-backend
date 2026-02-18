package data

import "log/slog"

type cache[T any] struct {
	value         Value[T]
	cached        T
	cachedVersion uint
}

func Cache[T any](value Value[T]) Value[T] {
	return &cache[T]{
		value:         value,
		cached:        value.Get(),
		cachedVersion: value.Version(),
	}
}

func (c *cache[T]) Get() T {
	if c.value.Version() != c.cachedVersion {
		slog.Debug("Refreshing cache", slog.Int("oldVersion", int(c.cachedVersion)), slog.Int("newVersion", int(c.value.Version())))
		c.cached = c.value.Get()
		c.cachedVersion = c.value.Version()
	}

	return c.cached
}

func (c *cache[T]) Version() uint {
	return c.value.Version()
}
