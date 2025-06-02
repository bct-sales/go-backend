package models

type ItemCategory struct {
	CategoryId Id
	Name       string
}

func IsValidCategoryName(name string) bool {
	return len(name) > 0
}
