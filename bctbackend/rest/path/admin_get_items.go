package path

type ItemsPath struct{}

func Items() *ItemsPath {
	return &ItemsPath{}
}

func (path *ItemsPath) String() string {
	return "/api/v1/items"
}
