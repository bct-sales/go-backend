package pdf

type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

func NewUniformInsets(value float64) *Insets {
	return &Insets{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}
