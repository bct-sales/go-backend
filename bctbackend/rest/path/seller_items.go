package path

import (
	"bctbackend/database/models"
	"fmt"
)

type SellerItemsPath struct{}

func SellerItems() *SellerItemsPath {
	return &SellerItemsPath{}
}

func (path *SellerItemsPath) WithRawSellerId(id string) string {
	return fmt.Sprintf("/api/v1/sellers/%s/items", id)
}

func (path *SellerItemsPath) WithSellerId(sellerId models.Id) string {
	return path.WithRawSellerId(models.IdToString(sellerId))
}
