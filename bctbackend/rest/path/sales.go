package path

import (
	"bctbackend/database/models"
	"fmt"
	"strconv"
	"strings"
)

type SalesPath struct {
	id              *string
	queryParameters map[string]string
}

func Sales() *SalesPath {
	return &SalesPath{}
}

func (path *SalesPath) WithQueryParameters(query string) string {
	return fmt.Sprintf("%s?%s", path.String(), query)
}

func (path *SalesPath) WithQueryParameter(key string, value string) *SalesPath {
	if path.queryParameters == nil {
		path.queryParameters = make(map[string]string)
	}

	path.queryParameters[key] = value

	return path
}

func (path *SalesPath) StartingAt(startId models.Id) *SalesPath {
	return path.WithQueryIdParameter("startId", startId)
}

func (path *SalesPath) WithLimitAndOffset(limit int, offset int) *SalesPath {
	return path.WithQueryIntParameter("limit", limit).WithQueryIntParameter("offset", offset)
}

func (path *SalesPath) AntiChronologically() *SalesPath {
	return path.WithQueryParameter("order", "antichronological")
}

func (path *SalesPath) WithQueryIntParameter(key string, value int) *SalesPath {
	return path.WithQueryParameter(key, strconv.Itoa(value))
}

func (path *SalesPath) WithQueryIdParameter(key string, id models.Id) *SalesPath {
	return path.WithQueryIntParameter(key, int(id.Int64()))
}

func (path *SalesPath) String() string {
	base := "/api/v1/sales"

	if path.id != nil {
		base = fmt.Sprintf("%s/%s", base, *path.id)
	}

	if len(path.queryParameters) > 0 {
		pairs := []string{}

		for key, value := range path.queryParameters {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}

		base = fmt.Sprintf("%s?%s", base, strings.Join(pairs, "&"))
	}

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
