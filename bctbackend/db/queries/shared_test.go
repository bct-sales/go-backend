package queries

import (
	database "bctbackend/db"
	models "bctbackend/db/models"
	"database/sql"
	"strconv"

	_ "modernc.org/sqlite"
)

func openDatabase() *sql.DB {
	db, error := sql.Open("sqlite", ":memory:")

	if error != nil {
		panic(error)
	}

	db.Exec("PRAGMA foreign_keys = 1")

	return db
}

func openInitializedDatabase() *sql.DB {
	db := openDatabase()

	database.InitializeDatabase(db)

	return db
}

func addSeller(db *sql.DB, id models.Id) {
	password := "test"

	AddUser(db, id, models.SellerRoleId, 0, password)
}

func addCashier(db *sql.DB, id models.Id) {
	password := "test"

	AddUser(db, id, models.CashierRoleId, 0, password)
}

func addItem(db *sql.DB, sellerId models.Id, index int) models.Id {
	var timestamp models.Timestamp = 0
	description := "description" + strconv.Itoa(index)

	var priceInCents models.MoneyInCents = 100 + int64(index)
	var itemCategoryId models.Id = 1
	donation := false
	charity := false

	itemId, err := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	return itemId
}
