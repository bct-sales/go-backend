package path

import (
	"bctbackend/database/models"
	"fmt"
)

type SalesItemsPath struct{}

func SalesItems() *SalesPath {
	return &SalesPath{}
}

func (path *SalesItemsPath) String() string {
	return "/api/v1/sales"
}

func (path *SalesPath) Raw(id string) string {
	return fmt.Sprintf("/api/v1/sales/items/%s", id)
}

func (path *SalesPath) Id(id models.Id) string {
	return path.Raw(models.IdToString(id))
}
