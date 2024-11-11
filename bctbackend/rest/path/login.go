package path

type loginPath struct{}

func Login() *loginPath {
	return &loginPath{}
}

func (path *loginPath) String() string {
	return "/api/v1/login"
}
