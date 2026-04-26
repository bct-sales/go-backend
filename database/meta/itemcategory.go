package meta

var ItemCategory = ItemCategoryMetadata{
	Table:          "item_categories",
	ItemCategoryID: "item_category_id",
	Name:           "name",
}

type ItemCategoryMetadata struct {
	Table          string
	ItemCategoryID string
	Name           string
}
