package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bctbackend/defs"
	restapi "bctbackend/rest/seller"

	models "bctbackend/database/models"
	"bctbackend/database/queries"

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
	t.Run("Successful", func(t *testing.T) {
		for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
			for _, price := range []models.MoneyInCents{1, 100, 10000} {
				for _, description := range []string{"Xyz", "Test Description"} {
					for _, categoryId := range defs.Categories() {
						for _, donation := range []bool{true, false} {
							for _, charity := range []bool{true, false} {
								t.Run(fmt.Sprintf("sellerId=%d price=%d description=%s categoryId=%d donation=%t charity=%t", sellerId, price, description, categoryId, donation, charity), func(t *testing.T) {
									db, router := createRestRouter()
									writer := httptest.NewRecorder()
									defer db.Close()

									seller := addTestSellerWithId(db, sellerId)
									sessionId := addTestSession(db, seller.UserId)

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

									if assert.NoError(t, err) {
										request.Header.Set("Content-Type", "application/json")
										request.AddCookie(createCookie(sessionId))

										if assert.NoError(t, err) {
											router.ServeHTTP(writer, request)

											if assert.Equal(t, http.StatusCreated, writer.Code) {
												response := fromJson[restapi.AddSellerItemResponse](writer.Body.String())

												itemsInDatabase, err := queries.GetItems(db)
												if assert.NoError(t, err) {
													if assert.Equal(t, 1, len(itemsInDatabase)) {
														itemInDatabase := itemsInDatabase[0]
														assert.Equal(t, response.ItemId, itemInDatabase.ItemId)
														assert.Equal(t, seller.UserId, itemInDatabase.SellerId)
														assert.Equal(t, price, itemInDatabase.PriceInCents)
														assert.Equal(t, description, itemInDatabase.Description)
														assert.Equal(t, categoryId, itemInDatabase.CategoryId)
														assert.Equal(t, donation, itemInDatabase.Donation)
														assert.Equal(t, charity, itemInDatabase.Charity)
													}
												}
											}
										}
									}
								})
							}
						}
					}
				}
			}
		}
	})

	t.Run("Failing", func(t *testing.T) {
		t.Run("Zero price", func(t *testing.T) {
			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := defs.Clothing50_56
			donation := false
			charity := false

			db, router := createRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := addTestSeller(db)
			sessionId := addTestSession(db, seller.UserId)

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

			if assert.NoError(t, err) {
				request.Header.Set("Content-Type", "application/json")
				request.AddCookie(createCookie(sessionId))

				if assert.NoError(t, err) {
					router.ServeHTTP(writer, request)

					assert.Equal(t, http.StatusBadRequest, writer.Code)
					itemsInDatabase, err := queries.GetItems(db)
					if assert.NoError(t, err) {
						assert.Equal(t, 0, len(itemsInDatabase))
					}
				}
			}
		})
	})

	t.Run("Invalid category", func(t *testing.T) {
		price := models.MoneyInCents(100)
		description := "Test Description"
		categoryId := models.NewId(1000)
		donation := false
		charity := false

		assert.NotContains(t, defs.Categories(), categoryId)

		db, router := createRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		seller := addTestSeller(db)
		sessionId := addTestSession(db, seller.UserId)

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

		if assert.NoError(t, err) {
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(createCookie(sessionId))

			if assert.NoError(t, err) {
				router.ServeHTTP(writer, request)

				assert.Equal(t, http.StatusBadRequest, writer.Code)
				itemsInDatabase, err := queries.GetItems(db)
				if assert.NoError(t, err) {
					assert.Equal(t, 0, len(itemsInDatabase))
				}
			}
		}
	})
}
