//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestListCategories(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()

	t.Run("Success", func(t *testing.T) {
		t.Run("As admin", func(t *testing.T) {
			t.Run("Without counts", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())

				url := path.Categories()
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.Len(t, actual.Categories, len(defaultCategoryNameTable))

				for _, category := range actual.Categories {
					require.Nil(t, category.Count)
				}
			})

			t.Run("With counts, including all items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())

				url := path.CategoriesWithCounts(queries.AllItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCategoriesSuccessResponse](t, writer.Body.String())

				for _, category := range actual.Categories {
					require.NotNil(t, category.Count)
					require.Equal(t, 0, *category.Count)
				}
			})

			t.Run("With sold item counts", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				categoryID1 := models.ID(1)
				categoryID2 := models.ID(2)
				categoryName1 := "foo"
				categoryName2 := "bar"

				_, sessionID := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				cashier := setup.Cashier()
				setup.Category(categoryID1, categoryName1)
				setup.Category(categoryID2, categoryName2)
				soldItem1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
				soldItem2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
				soldItem3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID2))
				setup.Sale(cashier.UserID, []models.ID{soldItem1.ItemID, soldItem2.ItemID})
				setup.Sale(cashier.UserID, []models.ID{soldItem3.ItemID})

				url := path.CategoriesWithSoldItemCounts()
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.Len(t, actual.Categories, 2)
				require.Equal(t, categoryID1, actual.Categories[0].CategoryID)
				require.Equal(t, categoryName1, actual.Categories[0].CategoryName)
				require.NotNil(t, actual.Categories[0].Count)
				require.Equal(t, 2, *actual.Categories[0].Count)
				require.Equal(t, categoryID2, actual.Categories[1].CategoryID)
				require.Equal(t, categoryName2, actual.Categories[1].CategoryName)
				require.NotNil(t, actual.Categories[1].Count)
				require.Equal(t, 1, *actual.Categories[1].Count)
			})
		})

		t.Run("As seller", func(t *testing.T) {
			t.Run("Without counts", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Seller())

				url := path.Categories()
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCategoriesSuccessResponse](t, writer.Body.String())

				for _, category := range actual.Categories {
					require.Nil(t, category.Count)
				}
			})

			t.Run("With counts", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Seller())

				url := path.CategoriesWithCounts(queries.AllItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})
		})

		t.Run("As cashier", func(t *testing.T) {
			t.Run("Without counts", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Cashier())

				url := path.Categories()
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCategoriesSuccessResponse](t, writer.Body.String())

				for _, category := range actual.Categories {
					require.Nil(t, category.Count)
				}
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("As cashier", func(t *testing.T) {
			t.Run("With counts", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Cashier())

				url := path.CategoriesWithCounts(queries.AllItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})
		})
	})
}
