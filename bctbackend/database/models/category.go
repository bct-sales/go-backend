package models

type ItemCategory struct {
	CategoryId Id
	Name       string
}

type ItemCategoryCount struct {
	CategoryId Id
	Name       string
	Count      int64
}

func NewCategory(
	id Id,
	name string) *ItemCategory {

	return &ItemCategory{
		CategoryId: id,
		Name:       name,
	}
}

func NewItemCategoryCount(
	id Id,
	name string,
	count int64) *ItemCategoryCount {

	return &ItemCategoryCount{
		CategoryId: id,
		Name:       name,
		Count:      count,
	}
}
