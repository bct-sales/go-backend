package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ListItemsItemData struct {
	ItemID       models.ID           `json:"itemId"`
	AddedAt      rest.DateTime       `json:"addedAt"`
	Description  string              `json:"description"`
	PriceInCents models.MoneyInCents `json:"priceInCents"`
	CategoryID   models.ID           `json:"categoryId"`
	SellerID     models.ID           `json:"sellerId"`
	Donation     bool                `json:"donation"`
	Charity      bool                `json:"charity"`
	Frozen       bool                `json:"frozen"`
	Large        bool                `json:"large"`
}

type ListItemsStatistics struct {
	TotalItemCount int                 `json:"totalItemCount"`
	TotalItemValue models.MoneyInCents `json:"totalItemValue"`
}

type ListItemsSuccessResponse struct {
	ListItemsStatistics
	Items []ListItemsItemData `json:"items"`
}

type listItemsParameters struct {
	category           *models.ID
	hidden             *bool
	descriptionPattern *string
	rowRange           queries.RowRange
}

func ListItems(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := &listItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type listItemsEndpoint struct {
	Endpoint
}

func (ep *listItemsEndpoint) execute(context *gin.Context) {
	if !ep.ensureUserHasCorrectRole(context) {
		return
	}

	parameters := ep.parseParameters(context)
	if parameters == nil {
		return
	}

	items, itemsOk := ep.fetchItemsFromDatabase(context, parameters)
	if !itemsOk {
		return
	}

	ep.sendSuccessResponse(context, items, parameters)
}

func (ep *listItemsEndpoint) parseParameters(context *gin.Context) *listItemsParameters {
	category, categoryOk := ep.parseCategoryQueryParameter(context)
	if !categoryOk {
		return nil
	}

	rowRange := ep.parseRowRangeQueryParameters(context)
	if rowRange == nil {
		return nil
	}

	hidden, hiddenOk := ep.parseHiddenQueryParameter(context)
	if !hiddenOk {
		return nil
	}

	descriptionPattern, descriptionPatternOk := ep.parseDescriptionQueryParameter(context)
	if !descriptionPatternOk {
		return nil
	}

	return &listItemsParameters{
		category:           category,
		rowRange:           *rowRange,
		hidden:             hidden,
		descriptionPattern: descriptionPattern,
	}
}

func (ep *listItemsEndpoint) parseDescriptionQueryParameter(context *gin.Context) (*string, bool) {
	parameterValue := context.Query("description")

	if parameterValue == "" {
		return nil, true
	} else {
		return &parameterValue, true
	}
}

func (ep *listItemsEndpoint) buildSQLQuery(parameters *listItemsParameters) *queries.GetItemsQuery {
	query := queries.NewGetItemsQuery()

	// Filtering based on category
	if parameters.category != nil {
		query.WithCategory(*parameters.category)
	}

	// Filtering based on visibility
	if parameters.hidden != nil {
		query.WithHidden(*parameters.hidden)
	}

	// Filtering based on description
	if parameters.descriptionPattern != nil {
		query.WithDescriptionPattern(fmt.Sprintf("%%%s%%", *parameters.descriptionPattern))
	}

	// Row range
	query.WithRowRange(&parameters.rowRange)

	return query
}

func (ep *listItemsEndpoint) parseCategoryQueryParameter(context *gin.Context) (*models.ID, bool) {
	logger := ep.Logger
	parameterValue := context.Query("category")

	// If no category query parameter is present, no filtering needs to occur
	if parameterValue == "" {
		return nil, true
	}

	value, err := strconv.ParseUint(parameterValue, 10, 64)
	if err != nil {
		logger.AddInformation("category", parameterValue)
		logger.InvalidInput("Invalid category parameter")

		failure_response.InvalidUriParameters(context, "invalid category identifier")

		return nil, false
	}

	categoryID := models.ID(value)
	return &categoryID, true
}

func (ep *listItemsEndpoint) ensureUserHasCorrectRole(context *gin.Context) bool {
	logger := ep.Logger

	if ep.RoleID != models.NewAdminRoleID() && ep.RoleID != models.NewCashierRoleID() {
		logger.InvalidRequest("Unauthorized access attempt to list all items")

		failure_response.WrongRole(context, "Only admins and cashiers can list all items")

		return false
	}

	return true
}

func (ep *listItemsEndpoint) fetchItemsFromDatabase(context *gin.Context, parameters *listItemsParameters) ([]*models.Item, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	var items []*models.Item

	query := ep.buildSQLQuery(parameters)
	if err := query.Execute(ctx, ep.Database, queries.CollectTo(&items)); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get items")

		failure_response.Unknown(context, "Failed to get items: "+err.Error())

		return nil, false
	}

	return items, true
}

func (ep *listItemsEndpoint) sendSuccessResponse(context *gin.Context, items []*models.Item, parameters *listItemsParameters) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendResponseAsJSON(context, items, parameters) },
		handleJSONFormatFunc:    func() { ep.sendResponseAsJSONFile(context, items) },
		handleCSVFormatFunc:     func() { ep.sendResponseAsCSVFile(context, items) },
	}
	ep.parseFormatQueryParameter(context, &formatHandler)
}

func (ep *listItemsEndpoint) sendResponseAsJSON(context *gin.Context, items []*models.Item, parameters *listItemsParameters) {
	itemsData := ep.convertData(items)

	itemStatistics, statsOk := ep.getItemStatistics(context, parameters)
	if !statsOk {
		return
	}

	response := ListItemsSuccessResponse{
		Items:               itemsData,
		ListItemsStatistics: *itemStatistics,
	}

	context.IndentedJSON(http.StatusOK, response)
}

func (ep *listItemsEndpoint) getItemStatistics(context *gin.Context, parameters *listItemsParameters) (*ListItemsStatistics, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	query := queries.NewGetItemStatisticsQuery()

	if parameters.hidden != nil {
		query.WithHidden(*parameters.hidden)
	}

	if parameters.category != nil {
		query.WithCategory(*parameters.category)
	}

	if parameters.descriptionPattern != nil {
		query.WithDescription(*parameters.descriptionPattern)
	}

	itemStatistics, err := query.Execute(ctx, ep.Database)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to count items")

		failure_response.Unknown(context, "Failed to count items: "+err.Error())

		return nil, false
	}

	result := ListItemsStatistics{
		TotalItemCount: itemStatistics.ItemCount,
		TotalItemValue: itemStatistics.TotalValueInCents,
	}

	return &result, true
}

func (ep *listItemsEndpoint) convertData(items []*models.Item) []ListItemsItemData {
	return algorithms.Map(items, func(item *models.Item) ListItemsItemData {
		return ListItemsItemData{
			ItemID:       item.ItemID,
			AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryID:   item.CategoryID,
			SellerID:     item.SellerID,
			Donation:     item.Donation,
			Charity:      item.Charity,
			Frozen:       item.Frozen,
			Large:        item.Large,
		}
	})
}

func (ep *listItemsEndpoint) sendResponseAsJSONFile(context *gin.Context, items []*models.Item) {
	convertedData := ep.convertData(items)

	context.Header("Content-Type", "application/json")
	context.Header("Content-Disposition", "attachment; filename=\"items.json\"")
	context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	context.Header("Pragma", "no-cache")
	context.IndentedJSON(http.StatusOK, convertedData)
}

func (ep *listItemsEndpoint) sendResponseAsCSVFile(context *gin.Context, items []*models.Item) {
	logger := ep.Logger
	ctx := context.Request.Context()
	categoryNameTable, err := queries.GetCategoryNameTable(ctx, ep.Database)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get category map")

		failure_response.Unknown(context, "Failed to get category map: "+err.Error())

		return
	}

	context.Header("Content-Type", "text/csv")
	context.Header("Content-Disposition", "attachment; filename=\"items.csv\"")
	context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	context.Header("Pragma", "no-cache")

	buffer := new(bytes.Buffer)
	if err := csv.FormatItemsAsCSV(items, categoryNameTable, buffer); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to format items as CSV")

		failure_response.Unknown(context, "Failed to format items as CSV: "+err.Error())

		return
	}
	string := buffer.String()
	context.String(http.StatusOK, string)
}
