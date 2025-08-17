package models

type ItemCategory struct {
	CategoryID Id
	Name       string
}

func IsValidCategoryName(name string) bool {
	return len(name) > 0
}
