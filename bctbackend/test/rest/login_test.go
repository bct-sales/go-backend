//go:build test

package rest

import (
	"bctbackend/database/models"
	"bctbackend/rest/path"
	"bctbackend/test"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
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
	}
}
