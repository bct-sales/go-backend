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

	_ "bctbackend/docs"
)

type GetItemsItemData struct {
	ItemId       models.Id           `json:"itemId"`
	AddedAt      rest.DateTime       `json:"addedAt"`
	Description  string              `json:"description"`
	PriceInCents models.MoneyInCents `json:"priceInCents"`
	CategoryId   models.Id           `json:"categoryId"`
	SellerId     models.Id           `json:"sellerId"`
	Donation     bool                `json:"donation"`
	Charity      bool                `json:"charity"`
	Frozen       bool                `json:"frozen"`
}

type GetItemsSuccessResponse struct {
	Items          []GetItemsItemData `json:"items"`
	TotalItemCount int                `json:"totalItemCount"`
}

// @Summary List all items of all sellers.
// @Description Returns all items of all sellers. Only accessible to users with the admin role.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} GetItemsSuccessResponse "Items successfully fetched"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch items"
// @Router /items [get]
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

	itemSelection := ep.parseItemSelectionQueryParameter()

	rowSelection := ep.parseRowSelectionQueryParameters()
	if rowSelection == nil {
		return
	}

	items, itemsOk := ep.fetchItemsFromDatabase(itemSelection, rowSelection)
	if !itemsOk {
		return
	}

	ep.sendSuccessResponse(items, itemSelection)
}

func (ep *listAllItemsEndpoint) ensureUserHasCorrectRole() bool {
	if ep.RoleId != models.NewAdminRoleId() {
		ep.Logger.InvalidRequest("Unauthorized access attempt to list all items")
		failure_response.WrongRole(ep.Context, "Only admins can list all items")
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

func (ep *listAllItemsEndpoint) fetchItemsFromDatabase(itemSelection queries.ItemSelection, rowSelection *queries.RowSelection) ([]*models.Item, bool) {
	items := []*models.Item{}

	if err := queries.GetItems(ep.Database, queries.CollectTo(&items), itemSelection, rowSelection); err != nil {
		ep.Logger.InternalError("Failed to get items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get items: "+err.Error())
		return nil, false
	}

	return items, true
}

func (ep *listAllItemsEndpoint) sendSuccessResponse(items []*models.Item, itemSelection queries.ItemSelection) {
	requestedFormat := ep.Context.Query("format")
	switch requestedFormat {
	case "":
		ep.sendResponseAsJSON(items, itemSelection)

	case "json":
		ep.sendResponseAsJSONFile(items)

	case "csv":
		ep.sendResponseAsCSVFile(items)

	default:
		ep.handleInvalidFormat(requestedFormat)
	}
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
	itemCount, err := queries.CountItems(ep.Database, itemSelection)
	if err != nil {
		ep.Logger.InternalError("Failed to count items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to count items: "+err.Error())
		return
	}

	response := GetItemsSuccessResponse{
		Items:          itemsData,
		TotalItemCount: itemCount,
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

func (ep *listAllItemsEndpoint) handleInvalidFormat(requestedFormat string) {
	ep.Logger.InvalidInput("Unknown format requested", "format", requestedFormat)
	failure_response.Unknown(ep.Context, "Unknown format: "+requestedFormat)
}
