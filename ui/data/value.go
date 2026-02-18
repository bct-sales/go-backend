package data

type Value[T any] interface {
	Get() T
	Version() uint
}
