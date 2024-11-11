package path

type LoginPath struct{}

func Login() *LoginPath {
	return &LoginPath{}
}

func (path *LoginPath) String() string {
	return "/api/v1/login"
}
