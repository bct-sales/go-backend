package path

type categoriesPath struct{}

func Categories() *categoriesPath {
	return &categoriesPath{}
}

func (path *categoriesPath) String() string {
	return "/api/v1/categories"
}
