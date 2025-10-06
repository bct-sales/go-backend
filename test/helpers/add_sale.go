//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
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

func AddSaleToDatabase(db *queries.TransactionalDatabaseQuerier, cashierID models.ID, itemIDs []models.ID, options ...func(*AddSaleData)) *models.Sale {
	data := AddSaleData{}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	saleID, err := queries.AddSale(db, cashierID, *data.TransactionTime, itemIDs)
	if err != nil {
		panic(err)
	}

	sale, err := queries.GetSaleWithID(db, saleID)
	if err != nil {
		panic(err)
	}

	return sale
}
