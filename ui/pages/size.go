package pages

type Size struct {
	Width  int
	Height int
}

func NewSize(width int, height int) *Size {
	return &Size{width, height}
}
