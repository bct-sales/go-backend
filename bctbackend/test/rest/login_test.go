//go:build test

package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/path"
	"bctbackend/security"
	"bctbackend/test"
	"bctbackend/test/setup"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuccessfulSellerLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	seller := setup.AddSellerToDatabase(db)

	form := url.Values{}
	form.Add("username", models.IdToString(seller.UserId))
	form.Add("password", seller.Password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusOK, writer.Code)

	var response map[string]string
	require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
	require.Equal(t, "seller", response["role"])

	cookies := writer.Result().Cookies()
	require.NotEmpty(t, cookies, "Expected cookies to be set")

	found := false
	sessionId := ""
	for _, cookie := range cookies {
		if cookie.Name == security.SessionCookieName {
			sessionId = cookie.Value
			found = true
			break
		}
	}
	require.True(t, found, "Expected session_id cookie to be set")

	sessionData, err := queries.GetSessionById(db, sessionId)
	require.NoError(t, err)
	require.Equal(t, seller.UserId, sessionData.UserId)
}

func TestSuccessfulAdminLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := setup.AddAdminToDatabase(db)

	form := url.Values{}
	form.Add("username", models.IdToString(admin.UserId))
	form.Add("password", admin.Password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusOK, writer.Code)

	var response map[string]string
	require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
	require.Equal(t, "admin", response["role"])

	cookies := writer.Result().Cookies()
	require.NotEmpty(t, cookies, "Expected cookies to be set")

	found := false
	sessionId := ""
	for _, cookie := range cookies {
		if cookie.Name == security.SessionCookieName {
			sessionId = cookie.Value
			found = true
			break
		}
	}
	require.True(t, found, "Expected session_id cookie to be set")

	sessionData, err := queries.GetSessionById(db, sessionId)
	require.NoError(t, err)
	require.Equal(t, admin.UserId, sessionData.UserId)
}

func TestSuccessfulCashierLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	cashier := setup.AddCashierToDatabase(db)

	form := url.Values{}
	form.Add("username", models.IdToString(cashier.UserId))
	form.Add("password", cashier.Password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusOK, writer.Code)

	var response map[string]string
	require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
	require.Equal(t, "cashier", response["role"])

	cookies := writer.Result().Cookies()
	require.NotEmpty(t, cookies, "Expected cookies to be set")

	found := false
	sessionId := ""
	for _, cookie := range cookies {
		if cookie.Name == security.SessionCookieName {
			sessionId = cookie.Value
			found = true
			break
		}
	}
	require.True(t, found, "Expected session_id cookie to be set")

	sessionData, err := queries.GetSessionById(db, sessionId)
	require.NoError(t, err)
	require.Equal(t, cashier.UserId, sessionData.UserId)
}

func TestUnknownUserLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	userId := models.Id(0)
	password := "xyz"

	form := url.Values{}
	form.Add("username", models.IdToString(userId))
	form.Add("password", password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "unknown_user")
}

func TestWrongPasswordLogin(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	seller := setup.AddSellerToDatabase(db)
	userId := seller.UserId
	password := "wrong password"

	require.NotEqual(t, password, seller.Password, "Bug in tests if this assertion fails")

	form := url.Values{}
	form.Add("username", models.IdToString(userId))
	form.Add("password", password)

	url := path.Login().String()
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "wrong_password")
}
