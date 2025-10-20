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

type ListSoldItemsEntry struct {
	SaleID          models.ID           `json:"saleId"`
	CashierID       models.ID           `json:"cashierId"`
	TransactionTime rest.DateTime       `json:"transactionTime"`
	ItemID          models.ID           `json:"itemId"`
	AddedAt         rest.DateTime       `json:"addedAt"`
	Description     string              `json:"description"`
	PriceInCents    models.MoneyInCents `json:"priceInCents"`
	ItemCategoryID  models.ID           `json:"itemCategory"`
	SellerID        models.ID           `json:"sellerId"`
	Donation        bool                `json:"donation"`
	Charity         bool                `json:"charity"`
}

type ListSoldItemsSuccessResponse struct {
	SoldItems []ListSoldItemsEntry `json:"soldItems"`
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
	logger := ep.Logger

	if ep.RoleID != models.NewAdminRoleID() {
		logger.InvalidRequest("Unauthorized access attempt to list sold items")
		failure_response.WrongRole(ep.Context, "Only admins can list sold items")
		return false
	}

	return true
}

func (ep *listSoldItemsEndpoint) fetchSoldItemsFromDatabase() ([]*queries.SoldItem, bool) {
	logger := ep.Logger

	query := ep.buildSQLQuery()
	soldItems, err := query.Execute(ep.Database)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get sold items")

		failure_response.Unknown(ep.Context, "Failed to get sold items: "+err.Error())

		return nil, false
	}

	return soldItems, true
}

func (ep *listSoldItemsEndpoint) buildSQLQuery() *queries.GetSoldItemsQuery {
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

	response := ListSoldItemsSuccessResponse{
		SoldItems: convertedData,
	}

	ep.Context.IndentedJSON(http.StatusOK, response)
}

func (ep *listSoldItemsEndpoint) sendResponseAsJSONFile(soldItems []*queries.SoldItem) {
	convertedData := ep.convertData(soldItems)

	ep.Context.Header("Content-Type", "application/json")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"sold-items.json\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")
	ep.Context.IndentedJSON(http.StatusOK, convertedData)
}

func (ep *listSoldItemsEndpoint) convertData(soldItems []*queries.SoldItem) []ListSoldItemsEntry {
	return algorithms.Map(soldItems, func(soldItem *queries.SoldItem) ListSoldItemsEntry {
		return ListSoldItemsEntry{
			SaleID:          soldItem.SaleID,
			CashierID:       soldItem.CashierID,
			TransactionTime: rest.ConvertTimestampToDateTime(soldItem.TransactionTime),
			ItemID:          soldItem.ItemID,
			AddedAt:         rest.ConvertTimestampToDateTime(soldItem.AddedAt),
			Description:     soldItem.Description,
			PriceInCents:    soldItem.PriceInCents,
			ItemCategoryID:  soldItem.ItemCategoryID,
			SellerID:        soldItem.SellerID,
			Donation:        soldItem.Donation,
			Charity:         soldItem.Charity,
		}
	})
}

func (ep *listSoldItemsEndpoint) sendResponseAsCSVFile(soldItems []*queries.SoldItem) {
	logger := ep.Logger

	categoryNameTable, err := queries.GetCategoryNameTable(ep.Database)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get category map")

		failure_response.Unknown(ep.Context, "Failed to get category map: "+err.Error())

		return
	}

	ep.Context.Header("Content-Type", "text/csv")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"sold-items.csv\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")

	buffer := new(bytes.Buffer)
	if err := csv.FormatSoldItemsAsCSV(soldItems, categoryNameTable, buffer); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to format items as CSV")

		failure_response.Unknown(ep.Context, "Failed to format items as CSV: "+err.Error())
		
		return
	}
	string := buffer.String()
	ep.Context.String(http.StatusOK, string)
}
