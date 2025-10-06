//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	. "bctbackend/test/setup"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestAddSellerItem(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()

	t.Run("Successful", func(t *testing.T) {
		for _, sellerID := range []models.ID{models.ID(1), models.ID(2)} {
			for _, price := range []models.MoneyInCents{1, 10000} {
				for _, description := range []string{"Xyz", "Test Description"} {
					for categoryID := range defaultCategoryNameTable {
						for _, donation := range []bool{true, false} {
							for _, charity := range []bool{true, false} {
								for _, delay := range []int{0, 100} {
									t.Run(fmt.Sprintf("sellerID=%d price=%d description=%s categoryID=%d donation=%t charity=%t", sellerID, price, description, categoryID, donation, charity), func(t *testing.T) {
										setup, router, writer := NewRestFixture(WithDefaultCategories)
										defer setup.Close()

										seller, sessionID := setup.LoggedIn(setup.Seller(aux.WithUserID(sellerID)), aux.WithExpiration(200))

										setup.Clock.Advance(models.Timestamp(delay))

										url := path.SellerItems(seller.UserID)
										payload := rest.AddSellerItemPayload{
											Price:       &price,
											Description: &description,
											CategoryID:  categoryID,
											Donation:    &donation,
											Charity:     &charity,
										}
										request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
										router.ServeHTTP(writer, request)

										require.Equal(t, http.StatusCreated, writer.Code)
										response := FromJSON[rest.AddSellerItemResponse](t, writer.Body.String())

										itemsInDatabase := []*models.Item{}
										err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
										require.NoError(t, err)
										require.Equal(t, 1, len(itemsInDatabase))

										itemInDatabase := itemsInDatabase[0]
										require.Equal(t, response.ItemID, itemInDatabase.ItemID)
										require.Equal(t, seller.UserID, itemInDatabase.SellerID)
										require.Equal(t, price, itemInDatabase.PriceInCents)
										require.Equal(t, description, itemInDatabase.Description)
										require.Equal(t, categoryID, itemInDatabase.CategoryID)
										require.Equal(t, donation, itemInDatabase.Donation)
										require.Equal(t, charity, itemInDatabase.Charity)
									})
								}
							}
						}
					}
				}
			}
		}
	})

	t.Run("Failing", func(t *testing.T) {
		t.Run("Zero price", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryID := aux.CategoryID_Clothing50_56
			donation := false
			charity := false

			seller, sessionID := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_price")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Empty description", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := ""
			categoryID := aux.CategoryID_Shoes
			donation := false
			charity := false

			seller, sessionID := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_item_description")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid category", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryID := models.ID(1000)
			donation := false
			charity := false

			require.NotContains(t, defaultCategoryNameTable, categoryID)

			seller, sessionID := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_category")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryID := aux.CategoryID_BabyChildEquipment
			donation := false
			charity := false

			seller := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Admin())

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryID := aux.CategoryID_Clothing104_116
			donation := false
			charity := false

			seller := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Cashier())
			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid url", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryID := aux.CategoryID_BabyChildEquipment
			donation := false
			charity := false

			_, sessionID := setup.LoggedIn(setup.Seller())

			url := path.SellerItemsStr("a")
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding as different seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryID := aux.CategoryID_BabyChildEquipment
			donation := false
			charity := false

			seller1 := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller1.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding item to nonexistent seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryID := aux.CategoryID_BabyChildEquipment
			donation := false
			charity := false

			_, sessionID := setup.LoggedIn(setup.Seller())
			nonexistentUserID := setup.GenerateNonexistentUserID(t)

			url := path.SellerItems(nonexistentUserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("No session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryID := aux.CategoryID_Clothing50_56
			donation := false
			charity := false

			seller := setup.Seller()

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryID := aux.CategoryID_Clothing50_56
			donation := false
			charity := false

			seller := setup.Seller()
			invalidSessionID := models.SessionID("xxx")

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(invalidSessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Expired session", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(50)
			description := "Test Description"
			categoryID := aux.CategoryID_Clothing50_56
			donation := false
			charity := false

			seller, sessionID := setup.LoggedIn(setup.Seller(), aux.WithExpiration(100))

			setup.Clock.Advance(100) // Advance time to ensure session is expired

			url := path.SellerItems(seller.UserID)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryID:  categoryID,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})
	})
}
