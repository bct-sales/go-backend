package data

type ConstantValue[T any] struct {
	value T
}

func NewConstant[T any](value T) *ConstantValue[T] {
	result := ConstantValue[T]{value: value}

	return &result
}

func (v *ConstantValue[T]) Get() T {
	return v.value
}

func (v *ConstantValue[T]) Version() uint {
	return 0
}
