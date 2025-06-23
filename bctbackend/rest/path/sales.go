package path

import (
	"bctbackend/database/models"
	"fmt"
	"strconv"
)

type SalesPath struct {
	Path
	id *string
}

func Sales() *SalesPath {
	return &SalesPath{}
}

func (path *SalesPath) StartingAt(startId models.Id) *SalesPath {
	path.WithQueryIdParameter("startId", startId)
	return path
}

func (path *SalesPath) WithLimitAndOffset(limit int, offset int) *SalesPath {
	path.WithQueryIntParameter("limit", limit).WithQueryIntParameter("offset", offset)
	return path
}

func (path *SalesPath) AntiChronologically() *SalesPath {
	path.WithQueryParameter("order", "antichronological")
	return path
}

func (path *SalesPath) String() string {
	base := "/api/v1/sales"

	if path.id != nil {
		base = fmt.Sprintf("%s/%s", base, *path.id)
	}

	base += path.QuerySuffixString()

	return base
}

func (path *SalesPath) Id(id models.Id) *SalesPath {
	s := strconv.FormatInt(id.Int64(), 10)
	path.id = &s

	return path
}

func (path *SalesPath) WithRawSaleId(id string) *SalesPath {
	path.id = &id
	return path
}
