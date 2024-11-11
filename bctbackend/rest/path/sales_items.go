package path

import (
	"bctbackend/database/models"
	"fmt"
)

type SalesItemsPath struct{}

func SalesItems() *SalesPath {
	return &SalesPath{}
}

func (path *SalesPath) WithRawItemId(id string) string {
	return fmt.Sprintf("/api/v1/sales/items/%s", id)
}

func (path *SalesPath) WithItemId(itemId models.Id) string {
	return path.WithRawItemId(models.IdToString(itemId))
}
