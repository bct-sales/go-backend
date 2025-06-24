package path

type logoutPath struct{}

func Logout() *logoutPath {
	return &logoutPath{}
}

func (path *logoutPath) String() string {
	return "/api/v1/logout"
}
