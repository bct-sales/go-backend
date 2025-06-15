package path

import (
	"bctbackend/database/models"
	"fmt"
)

type salesPath struct{}

func Sales() *salesPath {
	return &salesPath{}
}

func (path *salesPath) String() string {
	return "/api/v1/sales"
}

func (path *salesPath) Id(id models.Id) string {
	return path.WithRawSaleId(id.String())
}

func (path *salesPath) WithRawSaleId(id string) string {
	return fmt.Sprintf("/api/v1/sales/%s", id)
}
