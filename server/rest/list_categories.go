package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"fmt"
	"maps"
	"net/http"
	"slices"
)

type ListCategoriesSuccessResponse struct {
	Categories []CategoryData `json:"categories"`
}

type CategoryData struct {
	CategoryId   models.ID `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Count        *int      `json:"count,omitempty"`
}

func ListCategories(arguments *HandlerFunctionArguments) {
	endpoint := listCategoriesEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type listCategoriesEndpoint struct {
	Endpoint
}

func (ep *listCategoriesEndpoint) execute() {
	switch ep.Context.Query("counts") {
	case "all":
		ep.listCategoriesWithCounts(queries.AllItems)
		return

	case "hidden":
		ep.listCategoriesWithCounts(queries.OnlyHiddenItems)
		return

	case "visible":
		ep.listCategoriesWithCounts(queries.OnlyVisibleItems)
		return

	default:
		ep.listCategoriesWithoutCounts()
		return
	}
}

func (ep *listCategoriesEndpoint) listCategoriesWithCounts(itemSelection queries.ItemSelection) {
	context := ep.Context
	db := ep.Database
	roleId := ep.RoleID
	logger := ep.Logger

	if !roleId.IsAdmin() {
		logger.InvalidRequest("Unauthorized access to category counts")
		failure_response.WrongRole(context, "Only admins can access category counts")
		return
	}

	categoryCounts, err := queries.CountItemsPerCategory(db, itemSelection)
	if err != nil {
		logger.InternalError("Failed to fetch category counts", "error", err)
		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())
		return
	}

	categoryNameTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		logger.InternalError("Failed to fetch category name table", "error", err)
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
			logger.InvalidRequest("Unknown category ID", "categoryId", categoryId)
			failure_response.UnknownCategory(context, fmt.Sprintf("Unknown category ID %d", categoryId))
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

func (ep *listCategoriesEndpoint) listCategoriesWithoutCounts() {
	context := ep.Context
	db := ep.Database
	logger := ep.Logger

	categories, err := queries.GetCategories(db)
	if err != nil {
		logger.InternalError("Failed to fetch categories", "error", err)
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
