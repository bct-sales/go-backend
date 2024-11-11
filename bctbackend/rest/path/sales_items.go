package path

import (
	"bctbackend/database/models"
	"fmt"
)

type salesItemsPath struct{}

func SalesItems() *salesPath {
	return &salesPath{}
}

func (path *salesPath) WithRawItemId(id string) string {
	return fmt.Sprintf("/api/v1/sales/items/%s", id)
}

func (path *salesPath) WithItemId(itemId models.Id) string {
	return path.WithRawItemId(models.IdToString(itemId))
}
