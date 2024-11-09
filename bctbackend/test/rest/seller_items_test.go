package rest

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"

	"github.com/stretchr/testify/assert"
)

func TestListSellerItems(t *testing.T) {
	for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
		for _, itemCount := range []int{0, 1, 5, 100} {
			db, router := createRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := addTestSellerWithId(db, sellerId)
			sessionId := addTestSession(db, seller.UserId)

			expectedItems := []models.Item{}
			for i := 0; i < itemCount; i++ {
				expectedItems = append(expectedItems, *addTestItem(db, seller.UserId, i))
			}

			url := fmt.Sprintf("/api/v1/sellers/%d/items", seller.UserId)
			request, err := http.NewRequest("GET", url, nil)
			request.AddCookie(createCookie(sessionId))

			if assert.NoError(t, err) {
				router.ServeHTTP(writer, request)

				log.Println(writer.Body.String())
				if assert.Equal(t, http.StatusOK, writer.Code) {
					actual := fromJson[[]models.Item](writer.Body.String())
					assert.Equal(t, expectedItems, *actual)
				}
			}
		}
	}
}
