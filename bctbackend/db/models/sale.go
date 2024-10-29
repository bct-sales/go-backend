package models

type Sale struct {
	SaleId    Id
	CashierId Id
	Timestamp Timestamp
}

func NewSale(
	saleId Id,
	cashierId Id,
	timestamp Timestamp) *Sale {

	return &Sale{
		SaleId:    saleId,
		CashierId: cashierId,
		Timestamp: timestamp,
	}
}
