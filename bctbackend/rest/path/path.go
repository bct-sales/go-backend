package path

import (
	"bctbackend/database/models"
	"fmt"
	"strconv"
	"strings"
)

type Path[T any] struct {
	owner           T
	queryParameters map[string]string
}

func (path *Path[T]) WithQueryParameter(key string, value string) T {
	if path.queryParameters == nil {
		path.queryParameters = make(map[string]string)
	}

	path.queryParameters[key] = value

	return path.owner
}

func (path *Path[T]) WithQueryIntParameter(key string, value int) T {
	return path.WithQueryParameter(key, strconv.Itoa(value))
}

func (path *Path[T]) WithQueryIdParameter(key string, id models.Id) T {
	return path.WithQueryIntParameter(key, int(id.Int64()))
}

func (path *Path[T]) QuerySuffixString() string {
	if len(path.queryParameters) > 0 {
		pairs := []string{}

		for key, value := range path.queryParameters {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}

		return "?" + strings.Join(pairs, "&")
	}

	return ""
}

func (path *Path[T]) Limit(limit int) T {
	path.WithQueryIntParameter("limit", limit)
	return path.owner
}

func (path *Path[T]) Offset(offset int) T {
	path.WithQueryIntParameter("offset", offset)
	return path.owner
}
