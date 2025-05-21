package path

import "bctbackend/database/models"

type usersPath struct{}

func Users() *usersPath {
	return &usersPath{}
}

func (path *usersPath) String() string {
	return "/api/v1/users"
}

func (path *usersPath) WithUserId(userId models.Id) string {
	return path.WithRawUserId(userId.String())
}

func (path *usersPath) WithRawUserId(userId string) string {
	return "/api/v1/users/" + userId
}
