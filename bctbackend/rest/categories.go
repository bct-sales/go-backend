package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type ListCategoriesSuccessResponse struct {
	Counts []CategoryData `json:"counts"`
}

type CategoryData struct {
	CategoryId   models.Id `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Count        *int      `json:"count,omitempty"`
}

// @Summary Get number of items grouped by category.
// @Description Returns the number of items per category.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} ListCategoriesSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Unauthorized access"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch category counts"
// @Router /categories [get]
func ListCategories(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	includeCounts := context.Query("counts") == "true"

	if includeCounts {
		listCategoriesWithCounts(context, db, userId, roleId)
		return
	} else {
		listCategoriesWithoutCounts(context, db, userId, roleId)
		return
	}
}

func listCategoriesWithCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		slog.Error("Unauthorized access to category counts", "userId", userId, "roleId", roleId)
		failure_response.WrongRole(context, "Only admins can access category counts")
		return
	}

	categoryCounts, err := queries.GetCategoryCounts(db, false)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())
		return
	}

	categoryTable, err := queries.GetCategoryMap(db)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category table: "+err.Error())
		return
	}

	response := ListCategoriesSuccessResponse{
		Counts: []CategoryData{},
	}

	categoryIds := slices.Collect(maps.Keys(categoryCounts))
	slices.Sort(categoryIds)

	for _, categoryId := range categoryIds {
		categoryCount := categoryCounts[categoryId]
		categoryName, ok := categoryTable[categoryId]
		if !ok {
			failure_response.Unknown(context, fmt.Sprintf("Unknown category ID %d", categoryId))
			return
		}

		translatedCategoryCount := CategoryData{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			Count:        &categoryCount,
		}

		response.Counts = append(response.Counts, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}

func listCategoriesWithoutCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId && roleId != models.SellerRoleId {
		slog.Error("Unauthorized access to category counts", "userId", userId, "roleId", roleId)
		failure_response.WrongRole(context, "Only admins and sellers can access category names")
		return
	}

	categories, err := queries.GetCategories(db)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch categories: "+err.Error())
		return
	}

	response := ListCategoriesSuccessResponse{
		Counts: []CategoryData{},
	}

	for _, categoryCount := range categories {
		data := CategoryData{
			CategoryId:   categoryCount.CategoryId,
			CategoryName: categoryCount.Name,
		}

		response.Counts = append(response.Counts, data)
	}

	context.IndentedJSON(http.StatusOK, response)
}
