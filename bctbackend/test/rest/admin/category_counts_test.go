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
}
