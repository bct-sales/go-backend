//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	path "bctbackend/server/paths"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

type GetCategoriesSuccessResponse struct {
	Categories []struct {
		CategoryID   models.ID `json:"categoryId"`
		CategoryName string    `json:"categoryName"`
		Count        *int      `json:"count,omitempty"`
	} `json:"categories"`
}

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

				actual := FromJSON[GetCategoriesSuccessResponse](t, writer.Body.String())
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

				actual := FromJSON[GetCategoriesSuccessResponse](t, writer.Body.String())

				for _, category := range actual.Categories {
					require.NotNil(t, category.Count)
					require.Equal(t, 0, *category.Count)
				}
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

				actual := FromJSON[GetCategoriesSuccessResponse](t, writer.Body.String())

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

				actual := FromJSON[GetCategoriesSuccessResponse](t, writer.Body.String())

				for _, category := range actual.Categories {
					require.Nil(t, category.Count)
				}
			})

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
