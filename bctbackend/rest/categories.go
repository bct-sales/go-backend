package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type ListCategoriesSuccessResponse struct {
	Counts []CategoryData `json:"counts"`
}

type CategoryData struct {
	CategoryId   models.Id `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Count        *int64    `json:"count,omitempty"`
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
	switch roleId {
	case models.AdminRoleId:
		listCategoriesAsAdmin(context, db)
	case models.CashierRoleId:
		failure_response.WrongRole(context, "Cashiers are not allowed to list categories")
	case models.SellerRoleId:
		listCategoriesAsSeller(context, db, userId)
	}
}

func listCategoriesAsAdmin(context *gin.Context, db *sql.DB) {
	categoryCounts, err := queries.GetCategoryCounts(db)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())
		return
	}

	response := ListCategoriesSuccessResponse{
		Counts: []CategoryData{},
	}

	for _, categoryCount := range categoryCounts {
		translatedCategoryCount := CategoryData{
			CategoryId:   categoryCount.CategoryId,
			CategoryName: categoryCount.Name,
			Count:        &categoryCount.Count,
		}

		response.Counts = append(response.Counts, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}

func listCategoriesAsSeller(context *gin.Context, db *sql.DB, sellerId models.Id) {
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
