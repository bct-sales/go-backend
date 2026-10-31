package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
	"github.com/gin-gonic/gin"
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
	Large           bool                `json:"large"`
}

type ListSoldItemsSuccessResponse struct {
	SoldItems []ListSoldItemsEntry `json:"soldItems"`
}

type ListSoldItemsParameters struct{}

func ListSoldItems(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := &listSoldItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type listSoldItemsEndpoint struct {
	Endpoint
}

func (ep *listSoldItemsEndpoint) execute(context *gin.Context) {
	if !ep.ensureUserHasCorrectRole(context) {
		return
	}

	soldItems, fetchOk := ep.fetchSoldItemsFromDatabase(context)
	if !fetchOk {
		return
	}

	ep.sendSuccessResponse(context, soldItems)
}

func (ep *listSoldItemsEndpoint) ensureUserHasCorrectRole(context *gin.Context) bool {
	logger := ep.Logger

	if ep.RoleID != models.NewAdminRoleID() {
		logger.InvalidRequest("Unauthorized access attempt to list sold items")
		failure_response.WrongRole(context, "Only admins can list sold items")
		return false
	}

	return true
}

func (ep *listSoldItemsEndpoint) fetchSoldItemsFromDatabase(context *gin.Context) ([]*queries.SoldItem, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()

	query := ep.buildSQLQuery()
	soldItems, err := query.Execute(ctx, ep.Database)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get sold items")

		failure_response.Unknown(context, "Failed to get sold items: "+err.Error())

		return nil, false
	}

	return soldItems, true
}

func (ep *listSoldItemsEndpoint) buildSQLQuery() *queries.GetSoldItemsQuery {
	return queries.NewGetSoldItemsQuery()
}

func (ep *listSoldItemsEndpoint) sendSuccessResponse(context *gin.Context, items []*queries.SoldItem) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendResponseAsJSON(context, items) },
		handleJSONFormatFunc:    func() { ep.sendResponseAsJSONFile(context, items) },
		handleCSVFormatFunc:     func() { ep.sendResponseAsCSVFile(context, items) },
	}
	ep.parseFormatQueryParameter(context, &formatHandler)
}

func (ep *listSoldItemsEndpoint) sendResponseAsJSON(context *gin.Context, soldItems []*queries.SoldItem) {
	convertedData := ep.convertData(soldItems)

	response := ListSoldItemsSuccessResponse{
		SoldItems: convertedData,
	}

	context.IndentedJSON(http.StatusOK, response)
}

func (ep *listSoldItemsEndpoint) sendResponseAsJSONFile(context *gin.Context, soldItems []*queries.SoldItem) {
	convertedData := ep.convertData(soldItems)

	context.Header("Content-Type", "application/json")
	context.Header("Content-Disposition", "attachment; filename=\"sold-items.json\"")
	context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	context.Header("Pragma", "no-cache")
	context.IndentedJSON(http.StatusOK, convertedData)
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
			Large:           soldItem.Large,
		}
	})
}

func (ep *listSoldItemsEndpoint) sendResponseAsCSVFile(context *gin.Context, soldItems []*queries.SoldItem) {
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
	context.Header("Content-Disposition", "attachment; filename=\"sold-items.csv\"")
	context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	context.Header("Pragma", "no-cache")

	buffer := new(bytes.Buffer)
	if err := csv.FormatSoldItemsAsCSV(soldItems, categoryNameTable, buffer); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to format items as CSV")

		failure_response.Unknown(context, "Failed to format items as CSV: "+err.Error())

		return
	}
	string := buffer.String()
	context.String(http.StatusOK, string)
}
