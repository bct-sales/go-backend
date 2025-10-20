package models

import "bctbackend/algorithms"

type Item struct {
	ItemID       ID
	AddedAt      Timestamp
	Description  string
	PriceInCents MoneyInCents
	CategoryID   ID
	SellerID     ID
	Donation     bool
	Charity      bool
	Frozen       bool
	Hidden       bool
	Large        bool
}

func IsValidItemDescription(description string) bool {
	return len(description) > 0
}

// CollectItemIDs extracts the ItemID from each Item in the slice and returns a slice of ids.
func CollectItemIDs(items []*Item) []ID {
	return algorithms.Map(items, func(item *Item) ID { return item.ItemID })
}
