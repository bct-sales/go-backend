package models

type Sale struct {
	SaleId          Id
	CashierId       Id
	TransactionTime Timestamp
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
