package pdf

type Rectangle struct {
	Left   float64
	Top    float64
	Width  float64
	Height float64
}

func (r *Rectangle) Right() float64 {
	return r.Left + r.Width
}

func (r *Rectangle) Bottom() float64 {
	return r.Top + r.Height
}

func (r *Rectangle) Shrink(insets Insets) *Rectangle {
	return &Rectangle{
		Left:   r.Left + insets.Left,
		Top:    r.Top + insets.Top,
		Width:  r.Width - insets.Left - insets.Right,
		Height: r.Height - insets.Top - insets.Bottom,
	}
}
