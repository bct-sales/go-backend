package path

import (
	"bctbackend/database/models"
	"fmt"
)

type itemsPath struct{}

func Items() *itemsPath {
	return &itemsPath{}
}

func (path *itemsPath) String() string {
	return "/api/v1/items"
}

func (path *itemsPath) Id(id models.Id) string {
	return path.WithRawItemId(models.IdToString(id))
}

func (path *itemsPath) WithRawItemId(id string) string {
	return fmt.Sprintf("/api/v1/items/%s", id)
}
