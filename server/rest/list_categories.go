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
	CategoryID   models.ID `json:"categoryId"`
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

func (ep *listCategoriesEndpoint) execute() {
	parameters := ep.parseParameters()
	if parameters == nil {
		return
	}

	switch parameters.counts {
	case listCategoriesWithCountsOfAllItems:
		ep.listCategoriesWithCounts(queries.AllItems)
		return

	case listCategoriesWithCountsOfHiddenItems:
		ep.listCategoriesWithCounts(queries.OnlyHiddenItems)
		return

	case listCategoriesWithCountsOfVisibleItems:
		ep.listCategoriesWithCounts(queries.OnlyVisibleItems)
		return

	case listCategoriesWithCountsOfSoldItems:
		ep.listCategoriesWithSoldCounts()
		return

	case listCategoriesWithNoCounts:
		ep.listCategoriesWithoutCounts()
		return
	}
}

func (ep *listCategoriesEndpoint) parseParameters() *listCategoriesParameters {
	counts, countsOk := ep.parseCountsQueryParameter()
	if !countsOk {
		return nil
	}

	parameters := listCategoriesParameters{
		counts: counts,
	}

	return &parameters
}

func (ep *listCategoriesEndpoint) parseCountsQueryParameter() (listCategoriesCount, bool) {
	parameterValue := ep.Context.Query("counts")

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
		failure_response.InvalidUriParameters(ep.Context, "invalid value for counts query parameter")
		return 0, false
	}
}

func (ep *listCategoriesEndpoint) listCategoriesWithCounts(itemSelection queries.ItemSelection) {
	context := ep.Context
	db := ep.Database
	roleID := ep.RoleID
	logger := ep.Logger

	if !roleID.IsAdmin() {
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

	categoryIDs := slices.Collect(maps.Keys(categoryCounts))
	slices.Sort(categoryIDs)

	for _, categoryID := range categoryIDs {
		categoryCount := categoryCounts[categoryID]
		categoryName, ok := categoryNameTable[categoryID]
		if !ok {
			logger.InvalidRequest("Unknown category ID", "categoryID", categoryID)
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

func (ep *listCategoriesEndpoint) listCategoriesWithSoldCounts() {
	context := ep.Context
	db := ep.Database
	roleID := ep.RoleID
	logger := ep.Logger

	if !roleID.IsAdmin() {
		logger.InvalidRequest("Unauthorized access to category counts")
		failure_response.WrongRole(context, "Only admins can access category counts")
		return
	}

	categoryCounts, err := queries.CountSoldItemsPerCategory(db)
	if err != nil {
		logger.InternalError("Failed to fetch category counts of sold items", "error", err)
		failure_response.Unknown(context, "Failed to fetch sold item by category counts: "+err.Error())
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

	categoryIDs := slices.Collect(maps.Keys(categoryCounts))
	slices.Sort(categoryIDs)

	for _, categoryID := range categoryIDs {
		categoryCount := categoryCounts[categoryID]
		categoryName, ok := categoryNameTable[categoryID]
		if !ok {
			logger.InvalidRequest("Unknown category ID", "categoryID", categoryID)
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
			CategoryID:   categoryCount.CategoryID,
			CategoryName: categoryCount.Name,
			Count:        nil,
		}

		response.Categories = append(response.Categories, data)
	}

	context.IndentedJSON(http.StatusOK, response)
}
