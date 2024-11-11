package path

import "fmt"

type ItemsPath struct{}

func Items() *ItemsPath {
	return &ItemsPath{}
}

func (path *ItemsPath) String() string {
	return fmt.Sprintf("/api/v1/items")
}
