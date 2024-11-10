//go:build test

package test

import (
	database "bctbackend/database"
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/rest"
	"bctbackend/security"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	gin "github.com/gin-gonic/gin"
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

func CreateRestRouter() (*sql.DB, *gin.Engine) {
	db := OpenInitializedDatabase()
	router := rest.CreateRestRouter(db)

	return db, router
}

func AddTestUserWithId(db *sql.DB, id models.Id, roleId models.Id) models.User {
	password := "test"

	if err := queries.AddUserWithId(db, id, roleId, 0, password); err != nil {
		panic(err)
	}

	user, err := queries.GetUserWithId(db, id)

	if err != nil {
		panic(err)
	}

	return user
}

func AddTestUser(db *sql.DB, roleId models.Id) models.User {
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

func AddTestSeller(db *sql.DB) models.User {
	return AddTestUser(db, models.SellerRoleId)
}

func AddTestCashier(db *sql.DB) models.User {
	return AddTestUser(db, models.CashierRoleId)
}

func AddTestAdmin(db *sql.DB) models.User {
	return AddTestUser(db, models.AdminRoleId)
}

func AddTestAdminWithId(db *sql.DB, id models.Id) models.User {
	return AddTestUserWithId(db, id, models.AdminRoleId)
}

func AddTestSellerWithId(db *sql.DB, id models.Id) models.User {
	return AddTestUserWithId(db, id, models.SellerRoleId)
}

func AddTestCashierWithId(db *sql.DB, id models.Id) models.User {
	return AddTestUserWithId(db, id, models.CashierRoleId)
}

func AddTestItem(db *sql.DB, sellerId models.Id, index int) *models.Item {
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

func AddTestItemInCategory(db *sql.DB, sellerId models.Id, itemCategoryId models.Id) *models.Item {
	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
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

func AddTestSale(db *sql.DB, cashierId models.Id, itemIds []models.Id) models.Id {
	timestamp := models.NewTimestamp(0)

	saleId, err := queries.AddSale(db, cashierId, timestamp, itemIds)

	if err != nil {
		panic(err)
	}

	return saleId
}

func ToJson(x any) string {
	jsonData, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func FromJson[T any](jsonString string) *T {
	var x T
	err := json.Unmarshal([]byte(jsonString), &x)
	if err != nil {
		panic(err)
	}
	return &x
}

func AddTestSession(db *sql.DB, userId models.Id) string {
	sessionId, err := queries.AddSession(db, userId)

	if err != nil {
		panic(err)
	}

	return sessionId
}

func CreateCookie(sessionId string) *http.Cookie {
	return &http.Cookie{
		Name:     security.SessionCookieName,
		Value:    sessionId,
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	}
}
