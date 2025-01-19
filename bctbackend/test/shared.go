//go:build test

package test

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/rest"
	"bctbackend/security"
	"bctbackend/test/setup"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	gin "github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func CreateRestRouter() (*sql.DB, *gin.Engine) {
	db := setup.OpenInitializedDatabase()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rest.DefineEndpoints(db, router)

	return db, router
}

type AddUserData struct {
	UserId       *models.Id
	RoleId       models.Id
	Password     *string
	CreatedAt    *models.Timestamp
	LastActivity *models.Timestamp
}

func (data *AddUserData) FillWithDefaults() {
	if data.Password == nil {
		password := "test"
		data.Password = &password
	}

	if data.CreatedAt == nil {
		createdAt := models.NewTimestamp(0)
		data.CreatedAt = &createdAt
	}
}

func WithUserId(userId models.Id) func(*AddUserData) {
	return func(data *AddUserData) {
		data.UserId = &userId
	}
}

func WithPassword(password string) func(*AddUserData) {
	return func(data *AddUserData) {
		data.Password = &password
	}
}

func WithCreatedAt(createdAt models.Timestamp) func(*AddUserData) {
	return func(data *AddUserData) {
		data.CreatedAt = &createdAt
	}
}

func WithLastActivity(lastActivity models.Timestamp) func(*AddUserData) {
	return func(data *AddUserData) {
		data.LastActivity = &lastActivity
	}
}

func AddUserToDatabase(db *sql.DB, roleId models.Id, options ...func(*AddUserData)) models.User {
	data := AddUserData{
		RoleId: roleId,
	}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	var userId models.Id
	if data.UserId == nil {
		var err error
		userId, err = queries.AddUser(db, roleId, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	} else {
		userId = *data.UserId
		var err error
		err = queries.AddUserWithId(db, userId, roleId, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	}

	user, err := queries.GetUserWithId(db, userId)

	if err != nil {
		panic(err)
	}

	return user
}

func AddSellerToDatabase(db *sql.DB, options ...func(*AddUserData)) models.User {
	return AddUserToDatabase(db, models.SellerRoleId, options...)
}

func AddCashierToDatabase(db *sql.DB, options ...func(*AddUserData)) models.User {
	return AddUserToDatabase(db, models.CashierRoleId, options...)
}

func AddAdminToDatabase(db *sql.DB, options ...func(*AddUserData)) models.User {
	return AddUserToDatabase(db, models.AdminRoleId, options...)
}

type AddItemData struct {
	AddedAt      *models.Timestamp
	Description  *string
	PriceInCents *models.MoneyInCents
	ItemCategory *models.Id
	Donation     *bool
	Charity      *bool
}

func (data *AddItemData) FillWithDefaults() {
	if data.AddedAt == nil {
		addedAt := models.NewTimestamp(0)
		data.AddedAt = &addedAt
	}

	if data.Description == nil {
		description := "description"
		data.Description = &description
	}

	if data.PriceInCents == nil {
		priceInCents := models.NewMoneyInCents(100)
		data.PriceInCents = &priceInCents
	}

	if data.ItemCategory == nil {
		itemCategory := defs.Shoes
		data.ItemCategory = &itemCategory
	}

	if data.Donation == nil {
		donation := false
		data.Donation = &donation
	}

	if data.Charity == nil {
		charity := false
		data.Charity = &charity
	}
}

func WithAddedAt(addedAt models.Timestamp) func(*AddItemData) {
	return func(data *AddItemData) {
		data.AddedAt = &addedAt
	}
}

func WithDescription(description string) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Description = &description
	}
}

func WithPriceInCents(priceInCents models.MoneyInCents) func(*AddItemData) {
	return func(data *AddItemData) {
		data.PriceInCents = &priceInCents
	}
}

func WithItemCategory(itemCategory models.Id) func(*AddItemData) {
	return func(data *AddItemData) {
		data.ItemCategory = &itemCategory
	}
}

func WithDonation(donation bool) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Donation = &donation
	}
}

func WithCharity(charity bool) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Charity = &charity
	}
}

func WithDummyData(k int) func(*AddItemData) {
	return func(data *AddItemData) {
		addedAt := models.NewTimestamp(0)
		description := "description " + strconv.Itoa(k)
		priceInCents := models.NewMoneyInCents(100 + int64(k))
		itemCategory := defs.Shoes
		donation := k%2 == 0
		charity := k%3 == 0

		data.AddedAt = &addedAt
		data.Description = &description
		data.PriceInCents = &priceInCents
		if data.ItemCategory == nil {
			data.ItemCategory = &itemCategory
		}
		data.Donation = &donation
		data.Charity = &charity
	}
}

func AddItemToDatabase(db *sql.DB, sellerId models.Id, options ...func(*AddItemData)) *models.Item {
	data := AddItemData{}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	itemId, err := queries.AddItem(db, *data.AddedAt, *data.Description, *data.PriceInCents, *data.ItemCategory, sellerId, *data.Donation, *data.Charity)

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
