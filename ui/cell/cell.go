package cell

type Readable[T any] interface {
	Get() T
}

type Writable[T any] interface {
	Set(value T)
}

type Cell[T any] interface {
	Readable[T]
	Writable[T]
}

type Observable[T any] struct {
	value     T
	observers []func()
}

func NewObservable[T any](initialValue T) *Observable[T] {
	cell := Observable[T]{
		value:     initialValue,
		observers: nil,
	}

	return &cell
}

func (c *Observable[T]) Get() T {
	return c.value
}

func (c *Observable[T]) Set(value T) {
	c.value = value

	for _, observer := range c.observers {
		observer()
	}
}

func (c *Observable[T]) AddObserver(observer func()) {
	c.observers = append(c.observers, observer)
}
