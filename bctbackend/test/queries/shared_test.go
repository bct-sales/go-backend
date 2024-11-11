//go:build test

package queries

import (
	database "bctbackend/database"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"database/sql"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

func OpenDatabase() *sql.DB {
	db, error := sql.Open("sqlite", ":memory:")

	if error != nil {
		panic(error)
	}

	db.Exec("PRAGMA foreign_keys = 1")

	return db
}

func OpenInitializedDatabase() *sql.DB {
	db := OpenDatabase()

	if err := database.InitializeDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	return db
}

func AddUserWithId(db *sql.DB, id models.Id, roleId models.Id) {
	password := "test"

	queries.AddUserWithId(db, id, roleId, 0, password)
}

func AddUser(db *sql.DB, roleId models.Id) models.User {
	password := "test"

	userId, err := queries.AddUser(db, roleId, 0, password)

	if err != nil {
		panic(err)
	}

	user, err := queries.GetUserWithId(db, userId)

	if err != nil {
		panic(err)
	}

	return user
}

func AddSeller(db *sql.DB) models.User {
	return AddUser(db, models.SellerRoleId)
}

func AddCashier(db *sql.DB) models.User {
	return AddUser(db, models.CashierRoleId)
}

func AddAdmin(db *sql.DB) models.User {
	return AddUser(db, models.AdminRoleId)
}

func AddSellerWithId(db *sql.DB, id models.Id) {
	AddUserWithId(db, id, models.SellerRoleId)
}

func AddCashierWithId(db *sql.DB, id models.Id) {
	AddUserWithId(db, id, models.CashierRoleId)
}

func AddItem(db *sql.DB, sellerId models.Id, index int) *models.Item {
	timestamp := models.NewTimestamp(0)
	description := "description" + strconv.Itoa(index)
	priceInCents := models.NewMoneyInCents(100 + int64(index))
	itemCategoryId := defs.Shoes
	donation := false
	charity := false

	itemId, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	item, err := queries.GetItemWithId(db, itemId)

	if err != nil {
		panic(err)
	}

	return item
}

func AddItemInCategory(db *sql.DB, sellerId models.Id, itemCategoryId models.Id) models.Id {
	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	donation := false
	charity := false

	itemId, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	return itemId
}

func AddSaleToDatabase(db *sql.DB, cashierId models.Id, itemIds []models.Id) models.Id {
	timestamp := models.NewTimestamp(0)

	saleId, err := queries.AddSale(db, cashierId, timestamp, itemIds)

	if err != nil {
		panic(err)
	}

	return saleId
}
