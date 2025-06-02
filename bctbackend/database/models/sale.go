package models

type Sale struct {
	SaleId          Id
	CashierId       Id
	TransactionTime Timestamp
}

type SaleSummary struct {
	SaleId            Id
	CashierId         Id
	TransactionTime   Timestamp
	ItemCount         int
	TotalPriceInCents MoneyInCents
}
