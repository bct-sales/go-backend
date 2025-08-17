package paths

import (
	"bctbackend/database/models"
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

func (u *URL) WithQueryIdParameter(key string, id models.Id) *URL {
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

func (u *URL) AntiChronologically() *URL {
	return u.Order("antichronological")
}

func (u *URL) StartId(startId models.Id) *URL {
	return u.WithQueryIdParameter("startId", startId)
}
