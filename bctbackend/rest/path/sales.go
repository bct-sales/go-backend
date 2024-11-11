package path

type salesPath struct{}

func Sales() *salesPath {
	return &salesPath{}
}

func (path *salesPath) String() string {
	return "/api/v1/sales"
}
