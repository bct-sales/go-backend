package models

import "bctbackend/algorithms"

type Item struct {
	ItemID       Id
	AddedAt      Timestamp
	Description  string
	PriceInCents MoneyInCents
	CategoryID   Id
	SellerID     Id
	Donation     bool
	Charity      bool
	Frozen       bool
	Hidden       bool
}

func IsValidItemDescription(description string) bool {
	return len(description) > 0
}

// CollectItemIds extracts the ItemID from each Item in the slice and returns a slice of Ids.
func CollectItemIds(items []*Item) []Id {
	return algorithms.Map(items, func(item *Item) Id { return item.ItemID })
}
