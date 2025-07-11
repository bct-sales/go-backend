package models

type Sale struct {
	SaleID          Id
	CashierID       Id
	TransactionTime Timestamp
}

type SaleSummary struct {
	SaleID            Id
	CashierID         Id
	TransactionTime   Timestamp
	ItemCount         int
	TotalPriceInCents MoneyInCents
}
