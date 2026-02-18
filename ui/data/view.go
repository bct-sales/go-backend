package data

type View[T any] struct {
	value              Value[T]
	versionLastChecked uint
}

func NewView[T any](value Value[T]) View[T] {
	return View[T]{
		value:              value,
		versionLastChecked: value.Version(),
	}
}

func (view *View[T]) Get() T {
	view.versionLastChecked = view.value.Version()
	return view.value.Get()
}

func (view *View[T]) Updated() bool {
	return view.value.Version() != view.versionLastChecked
}

func (view *View[T]) Version() uint {
	return view.value.Version()
}
