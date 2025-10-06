package models

type ItemCategory struct {
	CategoryID ID
	Name       string
}

func IsValidCategoryName(name string) bool {
	return len(name) > 0
}
