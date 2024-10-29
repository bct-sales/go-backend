package models

type Item struct {
	ItemId       Id
	Timestamp    Timestamp
	Description  string
	PriceInCents MoneyInCents
	CategoryId   Id
	SellerId     Id
	Donation     bool
	Charity      bool
}

func NewItem(
	id Id,
	timestamp Timestamp,
	description string,
	priceInCents MoneyInCents,
	categoryId Id,
	sellerId Id,
	donation bool,
	charity bool) *Item {

	return &Item{
		ItemId:       id,
		Timestamp:    timestamp,
		Description:  description,
		PriceInCents: priceInCents,
		CategoryId:   categoryId,
		SellerId:     sellerId,
		Donation:     donation,
		Charity:      charity,
	}
}
