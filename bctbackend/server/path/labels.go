package path

type labelsPath struct{}

func Labels() *labelsPath {
	return &labelsPath{}
}

func (path *labelsPath) String() string {
	return "/api/v1/labels"
}
