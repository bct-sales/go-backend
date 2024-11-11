//go:build test

package queries

import (
	database "bctbackend/database"
	models "bctbackend/database/models"
	"bctbackend/defs"
	"database/sql"
	"log"
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

	if err := database.InitializeDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	return db
}

func addTestUserWithId(db *sql.DB, id models.Id, roleId models.Id) {
	password := "test"

	AddUserWithId(db, id, roleId, 0, password)
}

func addTestUser(db *sql.DB, roleId models.Id) models.User {
	password := "test"

	userId, err := AddUser(db, roleId, 0, password)

	if err != nil {
		panic(err)
	}

	user, err := GetUserWithId(db, userId)

	if err != nil {
		panic(err)
	}

	return user
}

func addTestSeller(db *sql.DB) models.User {
	return addTestUser(db, models.SellerRoleId)
}

func addTestCashier(db *sql.DB) models.User {
	return addTestUser(db, models.CashierRoleId)
}

func addTestAdmin(db *sql.DB) models.User {
	return addTestUser(db, models.AdminRoleId)
}

func addTestSellerWithId(db *sql.DB, id models.Id) {
	addTestUserWithId(db, id, models.SellerRoleId)
}

func addTestCashierWithId(db *sql.DB, id models.Id) {
	addTestUserWithId(db, id, models.CashierRoleId)
}

func addTestItem(db *sql.DB, sellerId models.Id, index int) *models.Item {
	timestamp := models.NewTimestamp(0)
	description := "description" + strconv.Itoa(index)
	priceInCents := models.NewMoneyInCents(100 + int64(index))
	itemCategoryId := defs.Shoes
	donation := false
	charity := false

	itemId, err := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	item, err := GetItemWithId(db, itemId)

	if err != nil {
		panic(err)
	}

	return item
}

func addTestItemInCategory(db *sql.DB, sellerId models.Id, itemCategoryId models.Id) models.Id {
	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	donation := false
	charity := false

	itemId, err := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	return itemId
}

func AddSaleToDatabase(db *sql.DB, cashierId models.Id, itemIds []models.Id) models.Id {
	timestamp := models.NewTimestamp(0)

	saleId, err := AddSale(db, cashierId, timestamp, itemIds)

	if err != nil {
		panic(err)
	}

	return saleId
}
