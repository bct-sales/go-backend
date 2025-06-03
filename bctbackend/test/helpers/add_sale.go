//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"database/sql"
)

type AddSaleData struct {
	TransactionTime *models.Timestamp
}

func WithTransactionTime(transactionTime models.Timestamp) func(*AddSaleData) {
	return func(data *AddSaleData) {
		data.TransactionTime = &transactionTime
	}
}

func (data *AddSaleData) FillWithDefaults() {
	if data.TransactionTime == nil {
		transactionTime := models.Timestamp(0)
		data.TransactionTime = &transactionTime
	}
}

func AddSaleToDatabase(db *sql.DB, cashierId models.Id, itemIds []models.Id, options ...func(*AddSaleData)) models.Id {
	data := AddSaleData{}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	saleId, err := queries.AddSale(db, cashierId, *data.TransactionTime, itemIds)

	if err != nil {
		panic(err)
	}

	return saleId
}
