package rest

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	db, _ := createRestRouter()
	defer db.Close()

	seller := addTestSeller(db)

	response, err := http.PostForm("/login", url.Values{
		"username": {strconv.Itoa(int(seller.UserId))},
		"password": {seller.Password},
	})

	if assert.NoError(t, err) {
		defer response.Body.Close()
	}
}
