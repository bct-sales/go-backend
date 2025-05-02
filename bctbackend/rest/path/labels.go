package path

import (
	"bctbackend/database/models"
	"fmt"
)

type labelsPath struct{}

func Labels() *labelsPath {
	return &labelsPath{}
}

func (path *labelsPath) WithRawSellerId(id string) string {
	return fmt.Sprintf("/api/v1/sellers/%s/labels", id)
}

func (path *labelsPath) WithSellerId(sellerId models.Id) string {
	return path.WithRawSellerId(models.IdToString(sellerId))
}
