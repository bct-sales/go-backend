//go:build test

package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/rest/path"
	"bctbackend/test"

	"github.com/stretchr/testify/assert"
)

func TestListSellerItems(t *testing.T) {
	for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
		for _, itemCount := range []int{0, 1, 5, 100} {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerWithIdToDatabase(db, sellerId)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			expectedItems := []models.Item{}
			for i := 0; i < itemCount; i++ {
				expectedItems = append(expectedItems, *test.AddItemToDatabase(db, seller.UserId, i))
			}

			url := path.SellerItems().Id(seller.UserId)
			request, err := http.NewRequest("GET", url, nil)
			request.AddCookie(test.CreateCookie(sessionId))

			if assert.NoError(t, err) {
				router.ServeHTTP(writer, request)

				if assert.Equal(t, http.StatusOK, writer.Code) {
					actual := test.FromJson[[]models.Item](writer.Body.String())
					assert.Equal(t, expectedItems, *actual)
				}
			}
		}
	}
}
