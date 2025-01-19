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
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func createSuccessResponse(countMap map[models.Id]int64) rest_admin.CategoryCountSuccessResponse {
	countArray := []rest_admin.CategoryCount{}

	for _, categoryId := range defs.ListCategories() {
		count, ok := countMap[categoryId]

		if !ok {
			count = 0
		}

		categoryName, err := defs.NameOfCategory(categoryId)

		if err != nil {
			panic(err)
		}

		countArray = append(countArray, rest_admin.CategoryCount{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			Count:        count,
		})
	}

	return rest_admin.CategoryCountSuccessResponse{Counts: countArray}
}

func TestCategoryCounts(t *testing.T) {
	t.Run("Zero items", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		admin := AddAdminToDatabase(db)
		sessionId := test.AddSessionToDatabase(db, admin.UserId)

		url := path.CategoryCounts().String()
		request := test.CreateGetRequest(url)
		request.AddCookie(test.CreateCookie(sessionId))

		router.ServeHTTP(writer, request)
		countMap := map[models.Id]int64{}
		expectedResponse := createSuccessResponse(countMap)
		actual := test.FromJson[rest_admin.CategoryCountSuccessResponse](writer.Body.String())
		require.Equal(t, expectedResponse, *actual)
	})

	for _, categoryId := range defs.ListCategories() {
		t.Run("Single item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := AddAdminToDatabase(db)
			seller := AddSellerToDatabase(db)
			AddItemToDatabase(db, seller.UserId, WithItemCategory(categoryId), WithDummyData(1))
			sessionId := test.AddSessionToDatabase(db, admin.UserId)

			url := path.CategoryCounts().String()
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)
			countMap := map[models.Id]int64{categoryId: 1}
			expected := createSuccessResponse(countMap)

			actual := test.FromJson[rest_admin.CategoryCountSuccessResponse](writer.Body.String())
			require.Equal(t, expected, *actual)
		})
	}

	for _, categoryId := range defs.ListCategories() {
		t.Run("Two items in same category", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := AddAdminToDatabase(db)
			seller := AddSellerToDatabase(db)
			AddItemToDatabase(db, seller.UserId, WithItemCategory(categoryId), WithDummyData(1))
			AddItemToDatabase(db, seller.UserId, WithItemCategory(categoryId), WithDummyData(1))
			sessionId := test.AddSessionToDatabase(db, admin.UserId)

			url := path.CategoryCounts().String()
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)
			countMap := map[models.Id]int64{categoryId: 2}
			expected := createSuccessResponse(countMap)

			actual := test.FromJson[rest_admin.CategoryCountSuccessResponse](writer.Body.String())
			require.Equal(t, expected, *actual)
		})
	}

	for _, categoryId1 := range defs.ListCategories() {
		for _, categoryId2 := range defs.ListCategories() {
			t.Run("Two items in potentially probably categories", func(t *testing.T) {
				db, router := test.CreateRestRouter()
				writer := httptest.NewRecorder()
				defer db.Close()

				admin := AddAdminToDatabase(db)
				seller := AddSellerToDatabase(db)
				AddItemToDatabase(db, seller.UserId, WithItemCategory(categoryId1), WithDummyData(1))
				AddItemToDatabase(db, seller.UserId, WithItemCategory(categoryId2), WithDummyData(2))
				sessionId := test.AddSessionToDatabase(db, admin.UserId)

				url := path.CategoryCounts().String()
				request := test.CreateGetRequest(url)
				request.AddCookie(test.CreateCookie(sessionId))

				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int64{categoryId1: 0, categoryId2: 0}
				countMap[categoryId1] += 1
				countMap[categoryId2] += 1
				expected := createSuccessResponse(countMap)

				actual := test.FromJson[rest_admin.CategoryCountSuccessResponse](writer.Body.String())
				require.Equal(t, expected, *actual)
			})
		}
	}
}
