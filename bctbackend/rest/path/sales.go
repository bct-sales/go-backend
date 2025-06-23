package path

import (
	"bctbackend/database/models"
	"strconv"
)

type SalesPath struct {
	Path[*SalesPath]
	id *string
}

func Sales() *SalesPath {
	//exhaustruct:ignore
	salesPath := SalesPath{}
	salesPath.owner = &salesPath
	return &salesPath
}

func (path *SalesPath) StartingAt(startId models.Id) *SalesPath {
	path.WithQueryIdParameter("startId", startId)
	return path
}

func (path *SalesPath) AntiChronologically() *SalesPath {
	path.WithQueryParameter("order", "antichronological")
	return path
}

func (path *SalesPath) String() string {
	base := "/api/v1/sales"

	if path.id != nil {
		base += "/" + *path.id
	}

	base += path.QuerySuffixString()

	return base
}

func (path *SalesPath) Id(id models.Id) *SalesPath {
	s := strconv.FormatInt(id.Int64(), 10)
	path.id = &s

	return path
}

func (path *SalesPath) IdStr(id string) *SalesPath {
	path.id = &id
	return path
}
