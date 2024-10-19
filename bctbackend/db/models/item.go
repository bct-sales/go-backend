package models

type Item struct {
	ItemId       Id
	Timestamp    Timestamp
	Description  string
	PriceInCents MoneyInCents
	CategoryId   Id
	OwnerId      Id
	RecipientId  Id
	Charity      bool
}

func NewItem(
	id Id,
	timestamp Timestamp,
	description string,
	priceInCents MoneyInCents,
	categoryId Id,
	ownerId Id,
	recipientId Id,
	charity bool) *Item {
	return &Item{
		ItemId:       id,
		Timestamp:    timestamp,
		Description:  description,
		PriceInCents: priceInCents,
		CategoryId:   categoryId,
		OwnerId:      ownerId,
		RecipientId:  recipientId,
		Charity:      charity,
	}
}
