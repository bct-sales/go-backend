package meta

var Sale = SaleMetadata{
	Table:           "sales",
	SaleID:          "sale_id",
	CashierID:       "cashier_id",
	TransactionTime: "transaction_time",
}

type SaleMetadata struct {
	Table           string
	SaleID          string
	CashierID       string
	TransactionTime string
}
