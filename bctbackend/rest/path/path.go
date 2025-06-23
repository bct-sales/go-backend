package path

import (
	"bctbackend/database/models"
	"fmt"
	"strconv"
	"strings"
)

type Path struct {
	queryParameters map[string]string
}

func (path *Path) WithQueryParameter(key string, value string) *Path {
	if path.queryParameters == nil {
		path.queryParameters = make(map[string]string)
	}

	path.queryParameters[key] = value

	return path
}

func (path *Path) WithQueryIntParameter(key string, value int) *Path {
	return path.WithQueryParameter(key, strconv.Itoa(value))
}

func (path *Path) WithQueryIdParameter(key string, id models.Id) *Path {
	return path.WithQueryIntParameter(key, int(id.Int64()))
}

func (path *Path) QuerySuffixString() string {
	if len(path.queryParameters) > 0 {
		pairs := []string{}

		for key, value := range path.queryParameters {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}

		return "?" + strings.Join(pairs, "&")
	}

	return ""
}
