package paths

import (
	"bctbackend/database/models"
	"net/url"
	"strconv"
	"strings"
)

type URL struct {
	pathSegments    []string
	queryParameters map[string]string
}

func NewURL() *URL {
	return &URL{
		pathSegments:    []string{},
		queryParameters: make(map[string]string),
	}
}

// AddPathSegment updates the URL object by adding an extra segment to its path.
// Returns the URL object so as to enable chaining.
func (u *URL) AddPathSegment(segment string) *URL {
	u.pathSegments = append(u.pathSegments, segment)
	return u
}

func (u *URL) AddQueryParameter(key, value string) *URL {
	u.queryParameters[key] = value
	return u
}

func (u *URL) String() string {
	path := "/" + strings.Join(u.pathSegments, "/")

	if len(u.queryParameters) > 0 {
		query := "?"
		for key, value := range u.queryParameters {
			query += key + "=" + value + "&"
		}
		query = query[:len(query)-1] // Remove the trailing '&'
		return path + query
	}
	return path
}

func (u *URL) AddQueryIDParameter(key string, id models.ID) *URL {
	return u.AddQueryParameter(key, id.String())
}

func (u *URL) AddQueryIntParameter(key string, value int) *URL {
	return u.AddQueryParameter(key, strconv.Itoa(value))
}

func (u *URL) AddLimit(limit int) *URL {
	return u.AddQueryIntParameter("limit", limit)
}

func (u *URL) AddOffset(offset int) *URL {
	return u.AddQueryIntParameter("offset", offset)
}

func (u *URL) AddOrder(order string) *URL {
	return u.AddQueryParameter("order", order)
}

func (u *URL) AddChronologicalOrder() *URL {
	return u.AddOrder("chronological")
}

func (u *URL) AddAntiChronologicalOrder() *URL {
	return u.AddOrder("antichronological")
}

func (u *URL) AddStartID(startID models.ID) *URL {
	return u.AddQueryIDParameter("startId", startID)
}

func (u *URL) AddCategoryFilter(categoryID models.ID) *URL {
	return u.AddQueryIDParameter("category", categoryID)
}

func (u *URL) AddHidden(value bool) *URL {
	return u.AddQueryParameter("hidden", strconv.FormatBool(value))
}

func (u *URL) AddDescription(description string) *URL {
	return u.AddQueryParameter("description", url.QueryEscape(description))
}
