//go:build test

package test

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"bctbackend/rest"
	"bctbackend/security"
	"bctbackend/test/setup"
	"database/sql"
	"encoding/json"
	"net/http"
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
