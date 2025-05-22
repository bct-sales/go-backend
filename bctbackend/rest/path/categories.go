package path

import "bctbackend/database/queries"

type categoriesPath struct{}

func Categories() *categoriesPath {
	return &categoriesPath{}
}

func (path *categoriesPath) String() string {
	return "/api/v1/categories"
}

func (path *categoriesPath) WithCounts(itemSelection queries.ItemSelection) string {
	switch itemSelection {
	case queries.AllItems:
		return path.String() + "?counts=all"
	case queries.OnlyHiddenItems:
		return path.String() + "?counts=hidden"
	case queries.OnlyVisibleItems:
		return path.String() + "?counts=visible"
	default:
		panic("bug: unknown item selection")
	}
}
