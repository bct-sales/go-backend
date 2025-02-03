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

func NewSale(
	saleId Id,
	cashierId Id,
	transactionTime Timestamp) *Sale {

	return &Sale{
		SaleId:          saleId,
		CashierId:       cashierId,
		TransactionTime: transactionTime,
	}
}

func NewSaleSummary(
	saleId Id,
	cashierId Id,
	transactionTime Timestamp,
	totalItems int,
	totalPrice MoneyInCents) *SaleSummary {

	return &SaleSummary{
		SaleId:            saleId,
		CashierId:         cashierId,
		TransactionTime:   transactionTime,
		ItemCount:         totalItems,
		TotalPriceInCents: totalPrice,
	}
}
