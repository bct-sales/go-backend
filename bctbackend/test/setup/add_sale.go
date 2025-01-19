//go:build test

package setup

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"database/sql"

	_ "modernc.org/sqlite"
)

func AddSaleToDatabase(db *sql.DB, cashierId models.Id, itemIds []models.Id) models.Id {
	transactionTime := models.NewTimestamp(0)

	return AddSaleAtTimeToDatabase(db, cashierId, itemIds, transactionTime)
}

func AddSaleAtTimeToDatabase(db *sql.DB, cashierId models.Id, itemIds []models.Id, transactionTime models.Timestamp) models.Id {
	saleId, err := queries.AddSale(db, cashierId, transactionTime, itemIds)

	if err != nil {
		panic(err)
	}

	return saleId
}
