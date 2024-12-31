//go:build test

package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/path"
	"bctbackend/security"
	"bctbackend/test"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulSellerLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	seller := test.AddSellerToDatabase(db)

	form := url.Values{}
	form.Add("username", models.IdToString(seller.UserId))
	form.Add("password", seller.Password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if assert.NoError(t, err) {
		router.ServeHTTP(writer, request)

		assert.Equal(t, http.StatusOK, writer.Code)

		cookies := writer.Result().Cookies()

		assert.NotEmpty(t, cookies, "Expected cookies to be set")
		found := false
		sessionId := ""
		for _, cookie := range cookies {
			if cookie.Name == security.SessionCookieName {
				sessionId = cookie.Value
				found = true
				break
			}
		}
		assert.True(t, found, "Expected session_id cookie to be set")

		sessionData, err := queries.GetSessionById(db, sessionId)

		if assert.NoError(t, err) {
			assert.Equal(t, seller.UserId, sessionData.UserId)
		}
	}
}

func TestSuccessfulAdminLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := test.AddAdminToDatabase(db)

	form := url.Values{}
	form.Add("username", models.IdToString(admin.UserId))
	form.Add("password", admin.Password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if assert.NoError(t, err) {
		router.ServeHTTP(writer, request)

		assert.Equal(t, http.StatusOK, writer.Code)

		cookies := writer.Result().Cookies()

		assert.NotEmpty(t, cookies, "Expected cookies to be set")
		found := false
		sessionId := ""
		for _, cookie := range cookies {
			if cookie.Name == security.SessionCookieName {
				sessionId = cookie.Value
				found = true
				break
			}
		}
		assert.True(t, found, "Expected session_id cookie to be set")

		sessionData, err := queries.GetSessionById(db, sessionId)

		if assert.NoError(t, err) {
			assert.Equal(t, admin.UserId, sessionData.UserId)
		}
	}
}

func TestSuccessfulCashierLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	cashier := test.AddCashierToDatabase(db)

	form := url.Values{}
	form.Add("username", models.IdToString(cashier.UserId))
	form.Add("password", cashier.Password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if assert.NoError(t, err) {
		router.ServeHTTP(writer, request)

		assert.Equal(t, http.StatusOK, writer.Code)

		cookies := writer.Result().Cookies()

		assert.NotEmpty(t, cookies, "Expected cookies to be set")
		found := false
		sessionId := ""
		for _, cookie := range cookies {
			if cookie.Name == security.SessionCookieName {
				sessionId = cookie.Value
				found = true
				break
			}
		}
		assert.True(t, found, "Expected session_id cookie to be set")

		sessionData, err := queries.GetSessionById(db, sessionId)

		if assert.NoError(t, err) {
			assert.Equal(t, cashier.UserId, sessionData.UserId)
		}
	}
}

func TestSessionExpiration(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := test.AddAdminToDatabase(db)
	sessionId := test.AddSessionToDatabaseWithExpiration(db, admin.UserId, -1)

	url := path.Items().String()
	request := test.CreateGetRequest(url)
	request.AddCookie(test.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)
	assert.Equal(t, http.StatusUnauthorized, writer.Code)
}
