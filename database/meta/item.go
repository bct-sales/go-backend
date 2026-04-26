package meta

var Item = ItemMetadata{
	Table:          "items",
	ItemID:         "item_id",
	AddedAt:        "added_at",
	Description:    "description",
	PriceInCents:   "price_in_cents",
	ItemCategoryID: "item_category_id",
	SellerID:       "seller_id",
	Donation:       "donation",
	Charity:        "charity",
	Frozen:         "frozen",
	Hidden:         "hidden",
	Large:          "large",
}

type ItemMetadata struct {
	Table          string
	ItemID         string
	AddedAt        string
	Description    string
	PriceInCents   string
	ItemCategoryID string
	SellerID       string
	Donation       string
	Charity        string
	Frozen         string
	Hidden         string
	Large          string
}
