package models

type Item struct {
	ItemID       Id
	AddedAt      Timestamp
	Description  string
	PriceInCents MoneyInCents
	CategoryID   Id
	SellerId     Id
	Donation     bool
	Charity      bool
	Frozen       bool
	Hidden       bool
}

func IsValidItemDescription(description string) bool {
	return len(description) > 0
}
