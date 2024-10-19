package clilib

type Parser[T any] struct {
	object *T
}

func NewParser[T any]() *Parser[T] {
	var empty T

	return &Parser[T]{
		object: &empty,
	}
}

type FlagHandler[T any] interface {
	initialize func(),
}

func (parser Parser[T]) Flag(name string, handler FlagHandler) {

}
