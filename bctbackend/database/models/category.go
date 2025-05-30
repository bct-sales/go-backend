package models

type ItemCategory struct {
	CategoryId Id
	Name       string
}

func NewCategory(id Id, name string) *ItemCategory {
	return &ItemCategory{
		CategoryId: id,
		Name:       name,
	}
}

func IsValidCategoryName(name string) bool {
	return len(name) > 0
}
