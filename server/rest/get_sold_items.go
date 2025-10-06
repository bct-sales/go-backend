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
)

type GetSoldItemsEntry struct {
	SaleId          models.Id           `json:"saleId"`
	CashierId       models.Id           `json:"cashierId"`
	TransactionTime rest.DateTime       `json:"transactionTime"`
	ItemId          models.Id           `json:"itemId"`
	AddedAt         rest.DateTime       `json:"addedAt"`
	Description     string              `json:"description"`
	PriceInCents    models.MoneyInCents `json:"priceInCents"`
	ItemCategoryID  models.Id           `json:"itemCategory"`
	SellerId        models.Id           `json:"sellerId"`
	Donation        bool                `json:"donation"`
	Charity         bool                `json:"charity"`
}

type GetSoldItemsSuccessResponse struct {
	SoldItems []GetSoldItemsEntry `json:"items"`
}

type ListSoldItemsParameters struct{}

func ListSoldItems(arguments *HandlerFunctionArguments) {
	endpoint := &listSoldItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type listSoldItemsEndpoint struct {
	Endpoint
}

func (ep *listSoldItemsEndpoint) execute() {
	if !ep.ensureUserHasCorrectRole() {
		return
	}

	soldItems, fetchOk := ep.fetchSoldItemsFromDatabase()
	if !fetchOk {
		return
	}

	ep.sendSuccessResponse(soldItems)
}

func (ep *listSoldItemsEndpoint) ensureUserHasCorrectRole() bool {
	if ep.RoleId != models.NewAdminRoleId() {
		ep.Logger.InvalidRequest("Unauthorized access attempt to list sold items")
		failure_response.WrongRole(ep.Context, "Only admins can list sold items")
		return false
	}

	return true
}

func (ep *listSoldItemsEndpoint) parseItemSelectionQueryParameter() queries.ItemSelection {
	switch ep.Context.Query("items") {
	case "all":
		return queries.AllItems
	case "hidden":
		return queries.OnlyHiddenItems
	default:
		return queries.OnlyVisibleItems
	}
}

func (ep *listSoldItemsEndpoint) fetchSoldItemsFromDatabase() ([]*queries.SoldItem, bool) {
	query := ep.buildSqlQuery()
	soldItems, err := query.Execute(ep.Database)

	if err != nil {
		ep.Logger.InternalError("Failed to get sold items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get sold items: "+err.Error())
		return nil, false
	}

	return soldItems, true
}

func (ep *listSoldItemsEndpoint) buildSqlQuery() *queries.GetSoldItemsQuery {
	return queries.NewGetSoldItemsQuery()
}

func (ep *listSoldItemsEndpoint) sendSuccessResponse(items []*queries.SoldItem) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendResponseAsJSON(items) },
		handleJSONFormatFunc:    func() { ep.sendResponseAsJSONFile(items) },
		handleCSVFormatFunc:     func() { ep.sendResponseAsCSVFile(items) },
	}
	ep.parseFormatQueryParameter(&formatHandler)
}

func (ep *listSoldItemsEndpoint) sendResponseAsJSON(soldItems []*queries.SoldItem) {
	convertedData := ep.convertData(soldItems)

	response := GetSoldItemsSuccessResponse{
		SoldItems: convertedData,
	}

	ep.Context.IndentedJSON(http.StatusOK, response)
}

func (ep *listSoldItemsEndpoint) sendResponseAsJSONFile(soldItems []*queries.SoldItem) {
	convertedData := ep.convertData(soldItems)

	ep.Context.Header("Content-Type", "application/json")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"items.json\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")
	ep.Context.IndentedJSON(http.StatusOK, convertedData)
}

func (ep *listSoldItemsEndpoint) convertData(soldItems []*queries.SoldItem) []GetSoldItemsEntry {
	return algorithms.Map(soldItems, func(soldItem *queries.SoldItem) GetSoldItemsEntry {
		return GetSoldItemsEntry{
			SaleId:          soldItem.SaleId,
			CashierId:       soldItem.CashierId,
			TransactionTime: rest.ConvertTimestampToDateTime(soldItem.TransactionTime),
			ItemId:          soldItem.ItemId,
			AddedAt:         rest.ConvertTimestampToDateTime(soldItem.AddedAt),
			Description:     soldItem.Description,
			PriceInCents:    soldItem.PriceInCents,
			ItemCategoryID:  soldItem.ItemCategoryId,
			SellerId:        soldItem.SellerId,
			Donation:        soldItem.Donation,
			Charity:         soldItem.Charity,
		}
	})
}

func (ep *listSoldItemsEndpoint) sendResponseAsCSVFile(soldItems []*queries.SoldItem) {
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
	if err := csv.FormatSoldItemsAsCSV(soldItems, categoryNameTable, buffer); err != nil {
		ep.Logger.InternalError("Failed to format items as CSV", "error", err)
		failure_response.Unknown(ep.Context, "Failed to format items as CSV: "+err.Error())
		return
	}
	string := buffer.String()
	ep.Context.String(http.StatusOK, string)
}
