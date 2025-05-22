//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	"bctbackend/rest/path"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

type GetCategoriesSuccessResponse struct {
	categories []struct {
		CategoryId   models.Id `json:"categoryId"`
		CategoryName string    `json:"categoryName"`
		Count        *int64    `json:"count,omitempty"`
	}
}

func TestGetCategories(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Without counts", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			url := path.Categories().String()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[GetCategoriesSuccessResponse](t, writer.Body.String())

			for _, category := range actual.categories {
				require.Nil(t, category.Count)
			}
		})

		t.Run("With counts", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			url := path.Categories().String()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[GetCategoriesSuccessResponse](t, writer.Body.String())

			for _, category := range actual.categories {
				require.NotNil(t, category.Count)
				require.Equal(t, int64(0), *category.Count)
			}
		})
	})
}
