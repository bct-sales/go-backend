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
	"strings"
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
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rest.DefineEndpoints(db, router)

	return db, router
}

func AddUserWithIdToDatabase(db *sql.DB, id models.Id, roleId models.Id) models.User {
	password := "test"
	createdAt := models.NewTimestamp(0)
	var lastActivity *models.Timestamp = nil

	if err := queries.AddUserWithId(db, id, roleId, createdAt, lastActivity, password); err != nil {
		panic(err)
	}

	user, err := queries.GetUserWithId(db, id)

	if err != nil {
		panic(err)
	}

	return user
}

func AddUserToDatabase(db *sql.DB, roleId models.Id) models.User {
	createdAt := models.NewTimestamp(0)
	var lastActivity *models.Timestamp = nil
	password := "test"

	userId, err := queries.AddUser(db, roleId, createdAt, lastActivity, password)

	if err != nil {
		panic(err)
	}

	user, err := queries.GetUserWithId(db, userId)

	if err != nil {
		panic(err)
	}

	return user
}

func AddSellerToDatabase(db *sql.DB) models.User {
	return AddUserToDatabase(db, models.SellerRoleId)
}

func AddCashierToDatabase(db *sql.DB) models.User {
	return AddUserToDatabase(db, models.CashierRoleId)
}

func AddAdminToDatabase(db *sql.DB) models.User {
	return AddUserToDatabase(db, models.AdminRoleId)
}

func AddAdminWithIdToDatabase(db *sql.DB, id models.Id) models.User {
	return AddUserWithIdToDatabase(db, id, models.AdminRoleId)
}

func AddSellerWithIdToDatabase(db *sql.DB, id models.Id) models.User {
	return AddUserWithIdToDatabase(db, id, models.SellerRoleId)
}

func AddCashierWithIdToDatabase(db *sql.DB, id models.Id) models.User {
	return AddUserWithIdToDatabase(db, id, models.CashierRoleId)
}

func AddItemToDatabase(db *sql.DB, sellerId models.Id, index int) *models.Item {
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

func AddItemInCategoryToDatabase(db *sql.DB, sellerId models.Id, itemCategoryId models.Id) *models.Item {
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

func AddSessionToDatabase(db *sql.DB, userId models.Id) string {
	return AddSessionToDatabaseWithExpiration(db, userId, 3600)
}

func AddSessionToDatabaseWithExpiration(db *sql.DB, userId models.Id, secondsBeforeExpiration int64) string {
	expirationTime := models.Now() + secondsBeforeExpiration
	sessionId, err := queries.AddSession(db, userId, expirationTime)

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

func CreateGetRequest(url string) *http.Request {
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	}

	return request
}

func CreatePostRequest[T any](url string, payload *T) *http.Request {
	payloadJson := ToJson(payload)
	request, err := http.NewRequest("POST", url, strings.NewReader(payloadJson))

	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json")

	return request
}
