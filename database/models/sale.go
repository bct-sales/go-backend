package models

type Sale struct {
	SaleID          ID
	CashierID       ID
	TransactionTime Timestamp
}

type SaleSummary struct {
	SaleID            ID
	CashierID         ID
	TransactionTime   Timestamp
	ItemCount         int
	TotalPriceInCents MoneyInCents
}
