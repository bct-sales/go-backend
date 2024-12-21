//go:build test

package rest

import (
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/defs"
	rest_admin "bctbackend/rest/admin"
	"bctbackend/rest/path"
	"bctbackend/test"

	"github.com/stretchr/testify/assert"
)

func createEmptyCategoryMap() map[models.Id]int64 {
	result := make(map[models.Id]int64)

	for _, categoryId := range defs.ListCategories() {
		result[categoryId] = 0
	}

	return result
}

func TestCategoryCounts(t *testing.T) {
	t.Run("Zero items", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		admin := test.AddAdminToDatabase(db)
		sessionId := test.AddSessionToDatabase(db, admin.UserId)

		url := path.CategoryCounts().String()
		request := test.CreateGetRequest(url)
		request.AddCookie(test.CreateCookie(sessionId))

		router.ServeHTTP(writer, request)
		expected := rest_admin.CategoryCountResponse{
			Counts: createEmptyCategoryMap(),
		}
		actual := test.FromJson[rest_admin.CategoryCountResponse](writer.Body.String())
		assert.Equal(t, expected, *actual)
	})

	for _, categoryId := range defs.ListCategories() {
		t.Run("Single item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := test.AddAdminToDatabase(db)
			seller := test.AddSellerToDatabase(db)
			test.AddItemInCategoryToDatabase(db, seller.UserId, categoryId)
			sessionId := test.AddSessionToDatabase(db, admin.UserId)

			url := path.CategoryCounts().String()
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)
			expected := rest_admin.CategoryCountResponse{
				Counts: createEmptyCategoryMap(),
			}
			expected.Counts[categoryId] = 1

			actual := test.FromJson[rest_admin.CategoryCountResponse](writer.Body.String())
			assert.Equal(t, expected, *actual)
		})
	}

	for _, categoryId := range defs.ListCategories() {
		t.Run("Two items in same category", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := test.AddAdminToDatabase(db)
			seller := test.AddSellerToDatabase(db)
			test.AddItemInCategoryToDatabase(db, seller.UserId, categoryId)
			test.AddItemInCategoryToDatabase(db, seller.UserId, categoryId)
			sessionId := test.AddSessionToDatabase(db, admin.UserId)

			url := path.CategoryCounts().String()
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)
			expected := rest_admin.CategoryCountResponse{
				Counts: createEmptyCategoryMap(),
			}
			expected.Counts[categoryId] = 2

			actual := test.FromJson[rest_admin.CategoryCountResponse](writer.Body.String())
			assert.Equal(t, expected, *actual)
		})
	}

	for _, categoryId1 := range defs.ListCategories() {
		for _, categoryId2 := range defs.ListCategories() {
			t.Run("Two items in potentially probably categories", func(t *testing.T) {
				db, router := test.CreateRestRouter()
				writer := httptest.NewRecorder()
				defer db.Close()

				admin := test.AddAdminToDatabase(db)
				seller := test.AddSellerToDatabase(db)
				test.AddItemInCategoryToDatabase(db, seller.UserId, categoryId1)
				test.AddItemInCategoryToDatabase(db, seller.UserId, categoryId2)
				sessionId := test.AddSessionToDatabase(db, admin.UserId)

				url := path.CategoryCounts().String()
				request := test.CreateGetRequest(url)
				request.AddCookie(test.CreateCookie(sessionId))

				router.ServeHTTP(writer, request)
				expected := rest_admin.CategoryCountResponse{
					Counts: createEmptyCategoryMap(),
				}
				expected.Counts[categoryId1] += 1
				expected.Counts[categoryId2] += 1

				actual := test.FromJson[rest_admin.CategoryCountResponse](writer.Body.String())
				assert.Equal(t, expected, *actual)
			})
		}
	}
}
