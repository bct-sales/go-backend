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
	Categories []CategoryData `json:"categories"`
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
func ListCategories(context *gin.Context, db *sql.DB, userId models.Id, roleId models.RoleId) {
	switch context.Query("counts") {
	case "all":
		listCategoriesWithCounts(context, db, userId, roleId, queries.AllItems)
		return

	case "hidden":
		listCategoriesWithCounts(context, db, userId, roleId, queries.OnlyHiddenItems)
		return

	case "visible":
		listCategoriesWithCounts(context, db, userId, roleId, queries.OnlyVisibleItems)
		return

	default:
		listCategoriesWithoutCounts(context, db, userId, roleId)
		return
	}
}

func listCategoriesWithCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.RoleId, itemSelection queries.ItemSelection) {
	if !roleId.IsAdmin() {
		slog.Error("Unauthorized access to category counts", "userId", userId, "roleId", roleId)
		failure_response.WrongRole(context, "Only admins can access category counts")
		return
	}

	categoryCounts, err := queries.GetCategoryCounts(db, itemSelection)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())
		return
	}

	categoryNameTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category table: "+err.Error())
		return
	}

	response := ListCategoriesSuccessResponse{
		Categories: []CategoryData{},
	}

	categoryIds := slices.Collect(maps.Keys(categoryCounts))
	slices.Sort(categoryIds)

	for _, categoryId := range categoryIds {
		categoryCount := categoryCounts[categoryId]
		categoryName, ok := categoryNameTable[categoryId]
		if !ok {
			failure_response.Unknown(context, fmt.Sprintf("Unknown category ID %d", categoryId))
			return
		}

		translatedCategoryCount := CategoryData{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			Count:        &categoryCount,
		}

		response.Categories = append(response.Categories, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}

func listCategoriesWithoutCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.RoleId) {
	if !roleId.IsAdmin() && !roleId.IsSeller() {
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
		Categories: []CategoryData{},
	}

	for _, categoryCount := range categories {
		data := CategoryData{
			CategoryId:   categoryCount.CategoryID,
			CategoryName: categoryCount.Name,
			Count:        nil,
		}

		response.Categories = append(response.Categories, data)
	}

	context.IndentedJSON(http.StatusOK, response)
}
