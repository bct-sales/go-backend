package path

type itemsPath struct{}

func Items() *itemsPath {
	return &itemsPath{}
}

func (path *itemsPath) String() string {
	return "/api/v1/admin/items"
}
