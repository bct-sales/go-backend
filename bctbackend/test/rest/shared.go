package rest

import (
	database "bctbackend/database"
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"bctbackend/rest"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

	gin "github.com/gin-gonic/gin"
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

func createRestRouter() (*sql.DB, *gin.Engine) {
	db := openInitializedDatabase()
	router := rest.CreateRestRouter(db)

	return db, router
}

func addTestUserWithId(db *sql.DB, id models.Id, roleId models.Id) {
	password := "test"

	queries.AddUserWithId(db, id, roleId, 0, password)
}

func addTestUser(db *sql.DB, roleId models.Id) models.Id {
	password := "test"

	userId, err := queries.AddUser(db, roleId, 0, password)

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
	itemCategoryId := models.Shoes
	donation := false
	charity := false

	itemId, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	if err != nil {
		panic(err)
	}

	return itemId
}

func addTestItemInCategory(db *sql.DB, sellerId models.Id, itemCategoryId models.Id) models.Id {
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

func addTestSale(db *sql.DB, cashierId models.Id, itemIds []models.Id) models.Id {
	timestamp := models.NewTimestamp(0)

	saleId, err := queries.AddSale(db, cashierId, timestamp, itemIds)

	if err != nil {
		panic(err)
	}

	return saleId
}

func toJson(x any) string {
	jsonData, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func fromJson[T any](jsonString string) *T {
	var x T
	err := json.Unmarshal([]byte(jsonString), &x)
	if err != nil {
		panic(err)
	}
	return &x
}
