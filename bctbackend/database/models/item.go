package models

type Item struct {
	ItemId       Id
	AddedAt      Timestamp
	Description  string
	PriceInCents MoneyInCents
	CategoryId   Id
	SellerId     Id
	Donation     bool
	Charity      bool
	Frozen       bool
	Hidden       bool
}

func NewItem(
	id Id,
	addedAt Timestamp,
	description string,
	priceInCents MoneyInCents,
	categoryId Id,
	sellerId Id,
	donation bool,
	charity bool,
	frozen bool,
	hidden bool) *Item {
	return &Item{
		ItemId:       id,
		AddedAt:      addedAt,
		Description:  description,
		PriceInCents: priceInCents,
		CategoryId:   categoryId,
		SellerId:     sellerId,
		Donation:     donation,
		Charity:      charity,
		Frozen:       frozen,
		Hidden:       hidden,
	}
}

func IsValidItemDescription(description string) bool {
	return len(description) > 0
}
