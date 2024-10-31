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

func addTestUserWithId(db *sql.DB, id models.Id, roleId models.Id) {
	password := "test"

	AddUserWithId(db, id, roleId, 0, password)
}

func addTestUser(db *sql.DB, roleId models.Id) models.Id {
	password := "test"

	userId, err := AddUser(db, roleId, 0, password)

	if err != nil {
		panic(err)
	}

	return userId
}

func addTestSeller(db *sql.DB) models.Id {
	return addTestUser(db, models.SellerRoleId)
}

func addTestCashier(db *sql.DB) models.Id {
	return addTestUser(db, models.CashierRoleId)
}

func addTestAdmin(db *sql.DB) models.Id {
	return addTestUser(db, models.AdminRoleId)
}

func addTestSellerWithId(db *sql.DB, id models.Id) {
	addTestUserWithId(db, id, models.SellerRoleId)
}

func addTestCashierWithId(db *sql.DB, id models.Id) {
	addTestUserWithId(db, id, models.CashierRoleId)
}

func addTestItem(db *sql.DB, sellerId models.Id, index int) models.Id {
	timestamp := models.NewTimestamp(0)
	description := "description" + strconv.Itoa(index)
	priceInCents := models.NewMoneyInCents(100 + int64(index))
	itemCategoryId := models.NewId(1)
	donation := false
	charity := false

	itemId, err := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	return itemId
}
