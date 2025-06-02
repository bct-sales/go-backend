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

func IsValidItemDescription(description string) bool {
	return len(description) > 0
}
