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

type GetItemsItemData struct {
	ItemId       models.ID           `json:"itemId"`
	AddedAt      rest.DateTime       `json:"addedAt"`
	Description  string              `json:"description"`
	PriceInCents models.MoneyInCents `json:"priceInCents"`
	CategoryId   models.ID           `json:"categoryId"`
	SellerId     models.ID           `json:"sellerId"`
	Donation     bool                `json:"donation"`
	Charity      bool                `json:"charity"`
	Frozen       bool                `json:"frozen"`
}

type GetItemsSuccessResponse struct {
	Items          []GetItemsItemData  `json:"items"`
	TotalItemCount int                 `json:"totalItemCount"`
	TotalItemValue models.MoneyInCents `json:"totalItemValue"`
}

func ListAllItems(arguments *HandlerFunctionArguments) {
	endpoint := &listAllItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type listAllItemsEndpoint struct {
	Endpoint
}

func (ep *listAllItemsEndpoint) execute() {
	if !ep.ensureUserHasCorrectRole() {
		return
	}

	items, itemsOk := ep.fetchItemsFromDatabase()
	if !itemsOk {
		return
	}

	// TODO Do not reparse the same parameters
	itemSelection := ep.parseItemSelectionQueryParameter()
	ep.sendSuccessResponse(items, itemSelection)
}

func (ep *listAllItemsEndpoint) buildSqlQuery() *queries.GetItemsQuery {
	query := queries.NewGetItemsQuery()

	if !ep.processQueryParameters(query) {
		return nil
	}

	return query
}

func (ep *listAllItemsEndpoint) processQueryParameters(query *queries.GetItemsQuery) bool {
	if !ep.processItemSelectionQueryParameter(query) {
		return false
	}

	if !ep.processRangeQueryParameters(query) {
		return false
	}

	if !ep.processCategoryQueryParameter(query) {
		return false
	}

	return true
}

func (ep *listAllItemsEndpoint) processCategoryQueryParameter(sqlQuery *queries.GetItemsQuery) bool {
	parameterValue := ep.Context.Query("category")

	if parameterValue != "" {
		categoryId, err := strconv.ParseUint(parameterValue, 10, 64)

		if err != nil {
			ep.Logger.InvalidInput("Invalid category parameter", "category", parameterValue)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Order must be 'antichronological'")
			return false
		}

		sqlQuery.WithCategory(models.ID(categoryId))
	}

	return true
}

func (ep *listAllItemsEndpoint) processItemSelectionQueryParameter(sqlQuery *queries.GetItemsQuery) bool {
	switch ep.Context.Query("items") {
	case "all":
		// NOP
	case "hidden":
		sqlQuery.WithHidden(true)
	case "visible":
		sqlQuery.WithHidden(false)
	default:
		sqlQuery.WithHidden(false)
	}

	return true
}

func (ep *listAllItemsEndpoint) processRangeQueryParameters(query *queries.GetItemsQuery) bool {
	optionalLimit, limitOk := ep.parseLimitQueryParameter()
	if !limitOk {
		return false
	}

	optionalOffset, offsetOk := ep.parseOffsetQueryParameter()
	if !offsetOk {
		return false
	}

	var limit uint64
	var offset uint64

	if optionalLimit == nil {
		limit = 1000000
	} else {
		limit = uint64(*optionalLimit)
	}

	if optionalOffset == nil {
		offset = 0
	} else {
		offset = uint64(*optionalOffset)
	}

	query.WithLimitAndOffset(limit, offset)

	return true
}

func (ep *listAllItemsEndpoint) ensureUserHasCorrectRole() bool {
	if ep.RoleId != models.NewAdminRoleId() && ep.RoleId != models.NewCashierRoleId() {
		ep.Logger.InvalidRequest("Unauthorized access attempt to list all items")
		failure_response.WrongRole(ep.Context, "Only admins and cashiers can list all items")
		return false
	}

	return true
}

func (ep *listAllItemsEndpoint) parseItemSelectionQueryParameter() queries.ItemSelection {
	switch ep.Context.Query("items") {
	case "all":
		return queries.AllItems
	case "hidden":
		return queries.OnlyHiddenItems
	default:
		return queries.OnlyVisibleItems
	}
}

func (ep *listAllItemsEndpoint) fetchItemsFromDatabase() ([]*models.Item, bool) {
	var items []*models.Item

	query := ep.buildSqlQuery()
	if err := query.Execute(ep.Database, queries.CollectTo(&items)); err != nil {
		ep.Logger.InternalError("Failed to get items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get items: "+err.Error())
		return nil, false
	}

	return items, true
}

func (ep *listAllItemsEndpoint) sendSuccessResponse(items []*models.Item, itemSelection queries.ItemSelection) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendResponseAsJSON(items, itemSelection) },
		handleJSONFormatFunc:    func() { ep.sendResponseAsJSONFile(items) },
		handleCSVFormatFunc:     func() { ep.sendResponseAsCSVFile(items) },
	}
	ep.parseFormatQueryParameter(&formatHandler)
}

func (ep *listAllItemsEndpoint) sendResponseAsJSON(items []*models.Item, itemSelection queries.ItemSelection) {
	itemsData := algorithms.Map(items, func(item *models.Item) GetItemsItemData {
		return GetItemsItemData{
			ItemId:       item.ItemID,
			AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryId:   item.CategoryID,
			SellerId:     item.SellerID,
			Donation:     item.Donation,
			Charity:      item.Charity,
			Frozen:       item.Frozen,
		}
	})
	itemStatistics, err := queries.GetItemStatistics(ep.Database, itemSelection)
	if err != nil {
		ep.Logger.InternalError("Failed to count items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to count items: "+err.Error())
		return
	}

	response := GetItemsSuccessResponse{
		Items:          itemsData,
		TotalItemCount: itemStatistics.ItemCount,
		TotalItemValue: itemStatistics.TotalValueInCents,
	}

	ep.Context.IndentedJSON(http.StatusOK, response)
}

func (ep *listAllItemsEndpoint) sendResponseAsJSONFile(items []*models.Item) {
	ep.Context.Header("Content-Type", "application/json")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"items.json\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")

	ep.Context.IndentedJSON(http.StatusOK, items)
}

func (ep *listAllItemsEndpoint) sendResponseAsCSVFile(items []*models.Item) {
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
