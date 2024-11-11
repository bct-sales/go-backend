package path

import (
	"bctbackend/database/models"
	"fmt"
)

type SellerItemsPath struct{}

func SellerItems() *SellerItemsPath {
	return &SellerItemsPath{}
}

func (path *SellerItemsPath) Raw(id string) string {
	return fmt.Sprintf("/api/v1/sellers/%s/items", id)
}

func (path *SellerItemsPath) Id(id models.Id) string {
	return path.Raw(models.IdToString(id))
}
