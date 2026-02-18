package data

type mappedValue[T any, R any] struct {
	argument    Value[T]
	transformer func(T) R
}

func MapValue[T any, R any](value Value[T], transformer func(T) R) Value[R] {
	result := mappedValue[T, R]{
		argument:    value,
		transformer: transformer,
	}

	return &result
}

func (value *mappedValue[T, R]) Get() R {
	return value.transformer(value.argument.Get())
}

func (value *mappedValue[T, R]) Version() uint {
	return value.argument.Version()
}

type mappedValue2[T1, T2, R any] struct {
	argument1   Value[T1]
	argument2   Value[T2]
	transformer func(T1, T2) R
}

func MapValue2[T1, T2, R any](value1 Value[T1], value2 Value[T2], transformer func(T1, T2) R) Value[R] {
	result := mappedValue2[T1, T2, R]{
		argument1:   value1,
		argument2:   value2,
		transformer: transformer,
	}

	return &result
}

func (value *mappedValue2[T1, T2, R]) Get() R {
	return value.transformer(value.argument1.Get(), value.argument2.Get())
}

func (value *mappedValue2[T1, T2, R]) Version() uint {
	return value.argument1.Version() + value.argument2.Version()
}

type mappedValue3[T1, T2, T3, R any] struct {
	argument1   Value[T1]
	argument2   Value[T2]
	argument3   Value[T3]
	transformer func(T1, T2, T3) R
}

func MapValue3[T1, T2, T3, R any](value1 Value[T1], value2 Value[T2], value3 Value[T3], transformer func(T1, T2, T3) R) Value[R] {
	result := mappedValue3[T1, T2, T3, R]{
		argument1:   value1,
		argument2:   value2,
		argument3:   value3,
		transformer: transformer,
	}

	return &result
}

func (value *mappedValue3[T1, T2, T3, R]) Get() R {
	return value.transformer(value.argument1.Get(), value.argument2.Get(), value.argument3.Get())
}

func (value *mappedValue3[T1, T2, T3, R]) Version() uint {
	return value.argument1.Version() + value.argument2.Version() + value.argument3.Version()
}
