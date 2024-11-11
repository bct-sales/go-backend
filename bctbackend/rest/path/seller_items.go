package path

import (
	"bctbackend/database/models"
	"fmt"
)

type sellerItemsPath struct{}

func SellerItems() *sellerItemsPath {
	return &sellerItemsPath{}
}

func (path *sellerItemsPath) WithRawSellerId(id string) string {
	return fmt.Sprintf("/api/v1/sellers/%s/items", id)
}

func (path *sellerItemsPath) WithSellerId(sellerId models.Id) string {
	return path.WithRawSellerId(models.IdToString(sellerId))
}
