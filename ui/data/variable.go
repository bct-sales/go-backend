package data

type Variable[T any] struct {
	value   T
	version uint
}

func NewVariable[T any](value T) Variable[T] {
	result := Variable[T]{
		value:   value,
		version: 0,
	}

	return result
}

func (v *Variable[T]) Set(value T) {
	v.value = value
	v.version++
}

func (v *Variable[T]) Get() T {
	return v.value
}

func (v *Variable[T]) Version() uint {
	return v.version
}
