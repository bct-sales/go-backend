package path

import "bctbackend/database/models"

type UserPath struct {
	Path[*UserPath]
}

func Users() *UserPath {
	//exhaustruct:ignore
	path := UserPath{}
	path.owner = &path
	return &path
}

func (path *UserPath) String() string {
	return "/api/v1/users"
}

func (path *UserPath) WithUserId(userId models.Id) string {
	return path.WithRawUserId(userId.String())
}

func (path *UserPath) WithRawUserId(userId string) string {
	return "/api/v1/users/" + userId
}
