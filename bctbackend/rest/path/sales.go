package path

import (
	"bctbackend/database/models"
	"fmt"
)

type salesPath struct{}

func Sales() *salesPath {
	return &salesPath{}
}

func (path *salesPath) WithQueryParameters(query string) string {
	return fmt.Sprintf("%s?%s", path.String(), query)
}

func (path *salesPath) StartingAt(startId models.Id) string {
	return path.WithQueryParameters(fmt.Sprintf("startId=%d", startId))
}

func (path *salesPath) WithLimitAndOffset(limit int, offset int) string {
	return path.WithQueryParameters(fmt.Sprintf("limit=%d&offset=%d", limit, offset))
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
