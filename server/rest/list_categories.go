package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"fmt"
	"github.com/gin-gonic/gin"
	"maps"
	"net/http"
	"slices"
)

type ListCategoriesSuccessResponse struct {
	Categories []CategoryData `json:"categories"`
}

type CategoryData struct {
	CategoryID   models.ID `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Count        *int      `json:"count,omitempty"`
}

func ListCategories(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := listCategoriesEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type listCategoriesEndpoint struct {
	Endpoint
}

type listCategoriesCount int

const (
	listCategoriesWithNoCounts listCategoriesCount = iota
	listCategoriesWithCountsOfVisibleItems
	listCategoriesWithCountsOfHiddenItems
	listCategoriesWithCountsOfAllItems
	listCategoriesWithCountsOfSoldItems
)

type listCategoriesParameters struct {
	counts listCategoriesCount
}

func (ep *listCategoriesEndpoint) execute(context *gin.Context) {
	parameters := ep.parseParameters(context)
	if parameters == nil {
		return
	}

	switch parameters.counts {
	case listCategoriesWithCountsOfAllItems:
		ep.listCategoriesWithCounts(context, queries.AllItems)
		return

	case listCategoriesWithCountsOfHiddenItems:
		ep.listCategoriesWithCounts(context, queries.OnlyHiddenItems)
		return

	case listCategoriesWithCountsOfVisibleItems:
		ep.listCategoriesWithCounts(context, queries.OnlyVisibleItems)
		return

	case listCategoriesWithCountsOfSoldItems:
		ep.listCategoriesWithSoldCounts(context)
		return

	case listCategoriesWithNoCounts:
		ep.listCategoriesWithoutCounts(context)
		return
	}
}

func (ep *listCategoriesEndpoint) parseParameters(context *gin.Context) *listCategoriesParameters {
	counts, countsOk := ep.parseCountsQueryParameter(context)
	if !countsOk {
		return nil
	}

	parameters := listCategoriesParameters{
		counts: counts,
	}

	return &parameters
}

func (ep *listCategoriesEndpoint) parseCountsQueryParameter(context *gin.Context) (listCategoriesCount, bool) {
	parameterValue := context.Query("counts")

	switch parameterValue {
	case "":
		return listCategoriesWithNoCounts, true
	case "visible":
		return listCategoriesWithCountsOfVisibleItems, true
	case "hidden":
		return listCategoriesWithCountsOfHiddenItems, true
	case "all":
		return listCategoriesWithCountsOfAllItems, true
	case "sold":
		return listCategoriesWithCountsOfSoldItems, true
	default:
		ep.Logger.InvalidInput("Unauthorized access to category counts")
		failure_response.InvalidUriParameters(context, "invalid value for counts query parameter")
		return 0, false
	}
}

func (ep *listCategoriesEndpoint) listCategoriesWithCounts(context *gin.Context, itemSelection queries.ItemSelection) {
	db := ep.Database
	roleID := ep.RoleID
	logger := ep.Logger
	ctx := context.Request.Context()

	if !roleID.IsAdmin() {
		logger.InvalidRequest("Unauthorized access to category counts")
		failure_response.WrongRole(context, "Only admins can access category counts")
		return
	}

	categoryCounts, err := queries.CountItemsPerCategory(ctx, db, itemSelection)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch category counts")

		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())

		return
	}

	categoryNameTable, err := queries.GetCategoryNameTable(ctx, db)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch category name table")

		failure_response.Unknown(context, "Failed to fetch category table: "+err.Error())

		return
	}

	response := ListCategoriesSuccessResponse{
		Categories: []CategoryData{},
	}

	categoryIDs := slices.Collect(maps.Keys(categoryCounts))
	slices.Sort(categoryIDs)

	for _, categoryID := range categoryIDs {
		categoryCount := categoryCounts[categoryID]
		categoryName, ok := categoryNameTable[categoryID]
		if !ok {
			logger.AddInformation("categoryID", categoryID)
			logger.InvalidRequest("Unknown category")

			failure_response.UnknownCategory(context, fmt.Sprintf("Unknown category ID %d", categoryID))

			return
		}

		translatedCategoryCount := CategoryData{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			Count:        &categoryCount,
		}

		response.Categories = append(response.Categories, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}

func (ep *listCategoriesEndpoint) listCategoriesWithSoldCounts(context *gin.Context) {
	db := ep.Database
	roleID := ep.RoleID
	logger := ep.Logger
	ctx := context.Request.Context()

	if !roleID.IsAdmin() {
		logger.InvalidRequest("Unauthorized access to category counts")

		failure_response.WrongRole(context, "Only admins can access category counts")

		return
	}

	categoryCounts, err := queries.CountSoldItemsPerCategory(ctx, db)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch category counts of sold items")

		failure_response.Unknown(context, "Failed to fetch sold item by category counts: "+err.Error())

		return
	}

	categoryNameTable, err := queries.GetCategoryNameTable(ctx, db)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch category name table")

		failure_response.Unknown(context, "Failed to fetch category table: "+err.Error())

		return
	}

	response := ListCategoriesSuccessResponse{
		Categories: []CategoryData{},
	}

	categoryIDs := slices.Collect(maps.Keys(categoryCounts))
	slices.Sort(categoryIDs)

	for _, categoryID := range categoryIDs {
		categoryCount := categoryCounts[categoryID]
		categoryName, ok := categoryNameTable[categoryID]
		if !ok {
			logger.AddInformation("categoryID", categoryID)
			logger.InvalidRequest("Unknown category ID")

			failure_response.UnknownCategory(context, fmt.Sprintf("Unknown category ID %d", categoryID))

			return
		}

		translatedCategoryCount := CategoryData{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			Count:        &categoryCount,
		}

		response.Categories = append(response.Categories, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}

func (ep *listCategoriesEndpoint) listCategoriesWithoutCounts(context *gin.Context) {
	db := ep.Database
	logger := ep.Logger
	ctx := context.Request.Context()

	categories, err := queries.GetCategories(ctx, db)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch categories")

		failure_response.Unknown(context, "Failed to fetch categories: "+err.Error())

		return
	}

	response := ListCategoriesSuccessResponse{
		Categories: []CategoryData{},
	}

	for _, categoryCount := range categories {
		data := CategoryData{
			CategoryID:   categoryCount.CategoryID,
			CategoryName: categoryCount.Name,
			Count:        nil,
		}

		response.Categories = append(response.Categories, data)
	}

	context.IndentedJSON(http.StatusOK, response)
}
