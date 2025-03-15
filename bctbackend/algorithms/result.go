package algorithms

type Result[T any] interface {
	GetValue() T
	IsFailure() error
}

type Success[T any] struct {
	value T
}

func (s Success[T]) GetValue() T {
	return s.value
}

func (s Success[T]) IsFailure() error {
	return nil
}

type Failure[T any] struct {
	err error
}

func (f Failure[T]) GetValue() T {
	panic("GetValue() called on a Failure result")
}

func (f Failure[T]) IsFailure() error {
	return f.err
}

func NewSuccess[T any](value T) Result[T] {
	return Success[T]{value}
}

func NewFailure[T any](err error) Result[T] {
	if err == nil {
		panic("NewFailure called with nil error")
	}

	return Failure[T]{err}
}
