package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	restapi "bctbackend/rest/seller"

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

				if assert.Equal(t, http.StatusOK, writer.Code) {
					actual := fromJson[[]models.Item](writer.Body.String())
					assert.Equal(t, expectedItems, *actual)
				}
			}
		}
	}
}

func TestAddSellerItem(t *testing.T) {
	for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
		db, router := createRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		seller := addTestSellerWithId(db, sellerId)
		sessionId := addTestSession(db, seller.UserId)

		price := models.MoneyInCents(100)
		donation := false
		description := "Test Description"
		categoryId := models.Clothing104_116
		charity := false
		payload := restapi.AddSellerItemPayload{
			Price:       price,
			Description: description,
			CategoryId:  categoryId,
			Donation:    &donation,
			Charity:     &charity,
		}

		payloadJson := toJson(payload)

		url := fmt.Sprintf("/api/v1/sellers/%d/items", seller.UserId)
		request, err := http.NewRequest("POST", url, strings.NewReader(payloadJson))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(createCookie(sessionId))

		if assert.NoError(t, err) {
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusCreated, writer.Code)
		}
	}
}
