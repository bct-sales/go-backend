package path

type usersPath struct{}

func Users() *usersPath {
	return &usersPath{}
}

func (path *usersPath) String() string {
	return "/api/v1/users"
}
