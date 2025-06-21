package path

import (
	"bctbackend/database/models"
	"fmt"
	"strings"
)

type itemsPath struct{}

func Items() *itemsPath {
	return &itemsPath{}
}

func (path *itemsPath) String() string {
	return "/api/v1/items"
}

func (path *itemsPath) Id(id models.Id) string {
	return path.WithRawItemId(id.String())
}

func (path *itemsPath) WithRawItemId(id string) string {
	return fmt.Sprintf("/api/v1/items/%s", id)
}

func (path *itemsPath) WithRowSelection(offset *int, limit *int) string {
	parts := []string{}

	if offset != nil {
		parts = append(parts, fmt.Sprintf("offset=%d", *offset))
	}

	if limit != nil {
		parts = append(parts, fmt.Sprintf("limit=%d", *limit))
	}

	joinedParts := strings.Join(parts, "&")
	url := "/api/v1/items"

	if joinedParts != "" {
		url += "?" + joinedParts
	}

	return url
}
