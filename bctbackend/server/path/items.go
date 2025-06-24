package path

import (
	"bctbackend/database/models"
)

type ItemsPath struct {
	Path[*ItemsPath]
	id *string
}

func Items() *ItemsPath {
	//exhaustruct:ignore
	path := ItemsPath{}
	path.owner = &path
	return &path
}

func (path *ItemsPath) String() string {
	base := "/api/v1/items"

	if path.id != nil {
		base += "/" + *path.id
	}

	base += path.QuerySuffixString()

	return base
}

func (path *ItemsPath) Id(id models.Id) *ItemsPath {
	s := id.String()
	path.id = &s
	return path
}

func (path *ItemsPath) IdStr(id string) *ItemsPath {
	path.id = &id
	return path
}
