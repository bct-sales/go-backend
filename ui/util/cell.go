package util

type ReadableCell[T any] interface {
	Get() T
}

type WritableCell[T any] interface {
	Set(value T)
}

type Cell[T any] interface {
	ReadableCell[T]
	WritableCell[T]
}

type ObservableCell[T any] struct {
	value     T
	observers []func()
}

func NewObservableCell[T any](initialValue T) *ObservableCell[T] {
	cell := ObservableCell[T]{
		value:     initialValue,
		observers: nil,
	}

	return &cell
}

func (c *ObservableCell[T]) Get() T {
	return c.value
}

func (c *ObservableCell[T]) Set(value T) {
	c.value = value

	for _, observer := range c.observers {
		observer()
	}
}

func (c *ObservableCell[T]) AddObserver(observer func()) {
	c.observers = append(c.observers, observer)
}
