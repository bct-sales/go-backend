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

func (r *Rectangle) ShrinkUniformly(amount float64) *Rectangle {
	return &Rectangle{
		Left:   r.Left + amount,
		Top:    r.Top + amount,
		Width:  r.Width - 2*amount,
		Height: r.Height - 2*amount,
	}
}
