package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
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
}

type ListItemsSuccessResponse struct {
	Items          []ListItemsItemData `json:"items"`
	TotalItemCount int                 `json:"totalItemCount"`
	TotalItemValue models.MoneyInCents `json:"totalItemValue"`
}

type listItemsParameters struct {
	category           *models.ID
	hidden             *bool
	descriptionPattern *string
	rowRange           queries.RowSelection
}

func ListItems(arguments *HandlerFunctionArguments) {
	endpoint := &listItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type listItemsEndpoint struct {
	Endpoint
}

func (ep *listItemsEndpoint) execute() {
	if !ep.ensureUserHasCorrectRole() {
		return
	}

	parameters := ep.parseParameters()
	if parameters == nil {
		return
	}

	items, itemsOk := ep.fetchItemsFromDatabase(parameters)
	if !itemsOk {
		return
	}

	ep.sendSuccessResponse(items, parameters)
}

func (ep *listItemsEndpoint) parseParameters() *listItemsParameters {
	category, categoryOk := ep.parseCategoryQueryParameter()
	if !categoryOk {
		return nil
	}

	rowRange := ep.parseRowSelectionQueryParameters()
	if rowRange == nil {
		return nil
	}

	hidden, hiddenOk := ep.parseHiddenQueryParameter()
	if !hiddenOk {
		return nil
	}

	descriptionPattern, descriptionPatternOk := ep.parseDescriptionQueryParameter()
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

func (ep *listItemsEndpoint) parseDescriptionQueryParameter() (*string, bool) {
	parameterValue := ep.Context.Query("description")

	if parameterValue == "" {
		return nil, true
	} else {
		return &parameterValue, true
	}
}

func (ep *listItemsEndpoint) buildSqlQuery(parameters *listItemsParameters) *queries.GetItemsQuery {
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
		query.WithDescriptionPattern(*parameters.descriptionPattern)
	}

	// Row range
	query.WithRowRange(&parameters.rowRange)

	return query
}

func (ep *listItemsEndpoint) parseCategoryQueryParameter() (*models.ID, bool) {
	parameterValue := ep.Context.Query("category")

	// If no category query parameter is present, no filtering needs to occur
	if parameterValue == "" {
		return nil, true
	}

	value, err := strconv.ParseUint(parameterValue, 10, 64)
	if err != nil {
		ep.Logger.InvalidInput("Invalid category parameter", "category", parameterValue)
		failure_response.InvalidUriParameters(ep.Context, "invalid category identifier")
		return nil, false
	}

	categoryID := models.ID(value)
	return &categoryID, true
}

func (ep *listItemsEndpoint) ensureUserHasCorrectRole() bool {
	if ep.RoleID != models.NewAdminRoleID() && ep.RoleID != models.NewCashierRoleID() {
		ep.Logger.InvalidRequest("Unauthorized access attempt to list all items")
		failure_response.WrongRole(ep.Context, "Only admins and cashiers can list all items")
		return false
	}

	return true
}

func (ep *listItemsEndpoint) fetchItemsFromDatabase(parameters *listItemsParameters) ([]*models.Item, bool) {
	var items []*models.Item

	query := ep.buildSqlQuery(parameters)
	if err := query.Execute(ep.Database, queries.CollectTo(&items)); err != nil {
		ep.Logger.InternalError("Failed to get items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get items: "+err.Error())
		return nil, false
	}

	return items, true
}

func (ep *listItemsEndpoint) sendSuccessResponse(items []*models.Item, parameters *listItemsParameters) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendResponseAsJSON(items, parameters) },
		handleJSONFormatFunc:    func() { ep.sendResponseAsJSONFile(items) },
		handleCSVFormatFunc:     func() { ep.sendResponseAsCSVFile(items) },
	}
	ep.parseFormatQueryParameter(&formatHandler)
}

func (ep *listItemsEndpoint) sendResponseAsJSON(items []*models.Item, parameters *listItemsParameters) {
	itemsData := ep.convertData(items)

	query := queries.NewGetItemStatisticsQuery()
	if parameters.hidden != nil {
		query.WithHidden(*parameters.hidden)
	}
	itemStatistics, err := query.Execute(ep.Database)
	if err != nil {
		ep.Logger.InternalError("Failed to count items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to count items: "+err.Error())
		return
	}

	response := ListItemsSuccessResponse{
		Items:          itemsData,
		TotalItemCount: itemStatistics.ItemCount,
		TotalItemValue: itemStatistics.TotalValueInCents,
	}

	ep.Context.IndentedJSON(http.StatusOK, response)
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
		}
	})
}

func (ep *listItemsEndpoint) sendResponseAsJSONFile(items []*models.Item) {
	convertedData := ep.convertData(items)

	ep.Context.Header("Content-Type", "application/json")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"items.json\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")
	ep.Context.IndentedJSON(http.StatusOK, convertedData)
}

func (ep *listItemsEndpoint) sendResponseAsCSVFile(items []*models.Item) {
	categoryNameTable, err := queries.GetCategoryNameTable(ep.Database)
	if err != nil {
		ep.Logger.InternalError("Failed to get category map", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get category map: "+err.Error())
		return
	}

	ep.Context.Header("Content-Type", "text/csv")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"items.csv\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")

	buffer := new(bytes.Buffer)
	if err := csv.FormatItemsAsCSV(items, categoryNameTable, buffer); err != nil {
		ep.Logger.InternalError("Failed to format items as CSV", "error", err)
		failure_response.Unknown(ep.Context, "Failed to format items as CSV: "+err.Error())
		return
	}
	string := buffer.String()
	ep.Context.String(http.StatusOK, string)
}
