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

// WithExtraPathSegment updates the URL object by adding an extra segment to its path.
// Returns the URL object so as to enable chaining.
func (u *URL) WithExtraPathSegment(segment string) *URL {
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

func (u *URL) WithQueryIDParameter(key string, id models.ID) *URL {
	return u.AddQueryParameter(key, id.String())
}

func (u *URL) WithQueryIntParameter(key string, value int) *URL {
	return u.AddQueryParameter(key, strconv.Itoa(value))
}

func (u *URL) Limit(limit int) *URL {
	return u.WithQueryIntParameter("limit", limit)
}

func (u *URL) Offset(offset int) *URL {
	return u.WithQueryIntParameter("offset", offset)
}

func (u *URL) Order(order string) *URL {
	return u.AddQueryParameter("order", order)
}

func (u *URL) Chronologically() *URL {
	return u.Order("chronological")
}

func (u *URL) AntiChronologically() *URL {
	return u.Order("antichronological")
}

func (u *URL) StartID(startID models.ID) *URL {
	return u.WithQueryIDParameter("startId", startID)
}

func (u *URL) CategoryFilter(categoryID models.ID) *URL {
	return u.WithQueryIDParameter("category", categoryID)
}

func (u *URL) Hidden(value bool) *URL {
	return u.AddQueryParameter("hidden", strconv.FormatBool(value))
}

func (u *URL) Description(description string) *URL {
	return u.AddQueryParameter("description", url.QueryEscape(description))
}
