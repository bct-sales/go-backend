package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"net/http"
	"strconv"

	_ "bctbackend/docs"
)

type ListSalesSaleData struct {
	SaleID            models.Id           `binding:"required" json:"saleId"`
	CashierID         models.Id           `binding:"required" json:"cashierId"`
	TransactionTime   rest.DateTime       `binding:"required" json:"transactionTime"`
	ItemCount         int                 `binding:"required" json:"itemCount"`
	TotalPriceInCents models.MoneyInCents `binding:"required" json:"totalPriceInCents"`
}

type ListSalesSuccessResponse struct {
	Sales                 []*ListSalesSaleData `json:"sales"`
	ItemCount             int                  `json:"itemCount"`
	DistinctSoldItemCount int                  `json:"distinctSoldItemCount"`
	TotalSoldItemCount    int                  `json:"totalSoldItemCount"`
	SaleCount             int                  `json:"saleCount"`
	TotalSaleValue        models.MoneyInCents  `json:"totalSaleValueInCents"`
}

type getSalesEndpoint struct {
	Endpoint
}

type getSalesQueryParameters struct {
	startId      *models.Id
	rowSelection *struct {
		limit  int
		offset int
	}
	orderedAntiChronologically bool
}

func GetSales(arguments *HandlerFunctionArguments) {
	endpoint := &getSalesEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(arguments.Database)
}

func (ep *getSalesEndpoint) execute(database Database) {
	if !ep.ensureUserIsAdmin() {
		return
	}

	queryParameters, queryParametersOk := ep.parseQueryParameters()
	if !queryParametersOk {
		return
	}

	response, responseOk := ep.fetchData(database, queryParameters)
	if !responseOk {
		return
	}

	ep.Context.IndentedJSON(http.StatusOK, response)
}

func (ep *getSalesEndpoint) fetchData(database Database, queryParameters *getSalesQueryParameters) (*ListSalesSuccessResponse, bool) {
	transaction, err := database.StartTransaction()
	if err != nil {
		ep.Logger.InternalError("Failed to create transaction", err)
		failure_response.Unknown(ep.Context, "Failed to create transaction: "+err.Error())
		return nil, false
	}
	defer transaction.Rollback()

	sales, getSalesOk := ep.getSales(transaction, queryParameters)
	if !getSalesOk {
		return nil, false
	}

	saleCount, countSalesOk := ep.countSales(transaction)
	if !countSalesOk {
		return nil, false
	}

	totalSaleValue, getTotalSalesValueOk := ep.getTotalSalesValue(transaction)
	if !getTotalSalesValueOk {
		return nil, false
	}

	itemCount, countItemsOk := ep.countItems(transaction)
	if !countItemsOk {
		return nil, false
	}

	distinctSoldItemCount, totalSoldItemCount, countSoldItemsOk := ep.countSoldItems(transaction)
	if !countSoldItemsOk {
		return nil, false
	}

	response := ListSalesSuccessResponse{
		Sales:                 sales,
		ItemCount:             itemCount,
		DistinctSoldItemCount: distinctSoldItemCount,
		TotalSoldItemCount:    totalSoldItemCount,
		SaleCount:             saleCount,
		TotalSaleValue:        totalSaleValue,
	}

	if err := transaction.Commit(); err != nil {
		// Unclear what to do, as only read operations were performed during the transaction
		ep.Logger.InternalError("Failed to commit transaction", "error", err)
		failure_response.Unknown(ep.Context, "Failed to commit transaction: "+err.Error())
		return nil, false
	}

	return &response, true
}

func (ep *getSalesEndpoint) countItems(transaction *queries.TransactionalDatabaseQuerier) (int, bool) {
	soldItemCount, err := queries.CountItems(transaction, queries.OnlyVisibleItems)
	if err != nil {
		ep.Logger.InternalError("Failed to get sold item count", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get sold item count: "+err.Error())
		return 0, false
	}
	return soldItemCount, true
}

func (ep *getSalesEndpoint) countSoldItems(transaction *queries.TransactionalDatabaseQuerier) (int, int, bool) {
	counts, err := queries.CountSoldItems(transaction)

	if err != nil {
		ep.Logger.InternalError("Failed to get sold item count", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get sold item count: "+err.Error())
		return 0, 0, false
	}

	return counts.Distinct, counts.IncludeMultiples, true
}

func (ep *getSalesEndpoint) countSales(transaction *queries.TransactionalDatabaseQuerier) (int, bool) {
	saleCount, err := queries.CountSales(transaction)

	if err != nil {
		ep.Logger.InternalError("Failed to get sales count", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get sales count: "+err.Error())
		return 0, false
	}

	return saleCount, true
}

func (ep *getSalesEndpoint) getTotalSalesValue(transaction *queries.TransactionalDatabaseQuerier) (models.MoneyInCents, bool) {
	totalValue, err := queries.GetTotalSalesValue(transaction)

	if err != nil {
		ep.Logger.InternalError("Failed to get total sales value", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get total sales value: "+err.Error())
		return 0, false
	}

	return totalValue, true

}

func (ep *getSalesEndpoint) ensureUserIsAdmin() bool {
	if ep.RoleId != models.NewAdminRoleId() {
		ep.Logger.InvalidRequest("Unauthorized access to list all sales", "userId", ep.UserId, "roleId", ep.RoleId)
		failure_response.WrongRole(ep.Context, "Only admins can list all items")
		return false
	}

	return true
}

func (ep *getSalesEndpoint) getSales(transaction *queries.TransactionalDatabaseQuerier, queryParameters *getSalesQueryParameters) ([]*ListSalesSaleData, bool) {
	sales := make([]*ListSalesSaleData, 0, 25)
	processSale := func(sale *models.SaleSummary) error {
		saleData := ListSalesSaleData{
			SaleID:            sale.SaleID,
			CashierID:         sale.CashierID,
			TransactionTime:   rest.ConvertTimestampToDateTime(sale.TransactionTime),
			ItemCount:         sale.ItemCount,
			TotalPriceInCents: sale.TotalPriceInCents,
		}

		sales = append(sales, &saleData)
		return nil
	}

	query := ep.buildQuery(queryParameters)

	if err := query.Execute(transaction, processSale); err != nil {
		ep.Logger.InternalError("Failed to get sales", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get sales: "+err.Error())
		return nil, false
	}

	return sales, true
}

func (ep *getSalesEndpoint) buildQuery(queryParameters *getSalesQueryParameters) *queries.GetSalesQuery {
	query := queries.NewGetSalesQuery()

	if queryParameters.startId != nil {
		query.WithIdGreaterThanOrEqualTo(*queryParameters.startId)
	}

	if queryParameters.rowSelection != nil {
		query.WithRowSelection(queryParameters.rowSelection.limit, queryParameters.rowSelection.offset)
	}

	if queryParameters.orderedAntiChronologically {
		query.OrderedAntiChronologically()
	}

	return query
}

func (ep *getSalesEndpoint) parseQueryParameters() (*getSalesQueryParameters, bool) {
	startId, ok := ep.parseStartId()
	if !ok {
		return nil, false
	}

	rowSelection, ok := ep.parseRowSelection()
	if !ok {
		return nil, false
	}

	order, ok := ep.parseOrder()
	if !ok {
		return nil, false
	}

	queryParameters := getSalesQueryParameters{
		startId:                    startId,
		rowSelection:               rowSelection,
		orderedAntiChronologically: order,
	}

	return &queryParameters, true
}

func (ep *getSalesEndpoint) parseStartId() (*models.Id, bool) {
	if startIdStr, exists := ep.Context.GetQuery("startId"); exists {
		startId, err := models.ParseId(startIdStr)
		if err != nil {
			ep.Logger.InvalidInput("Failed to parse startId parameter", "startId", startIdStr, "error", err)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Invalid startId parameter: "+err.Error())
			return nil, false
		}
		return &startId, true
	}

	return nil, true
}

func (ep *getSalesEndpoint) parseRowSelection() (*struct {
	limit  int
	offset int
}, bool) {

	limitString, limitExists := ep.Context.GetQuery("limit")
	offsetString, offsetExists := ep.Context.GetQuery("offset")

	if !limitExists && !offsetExists {
		return nil, true
	}

	if limitExists && !offsetExists {
		offsetString = "0" // Default offset to 0 if limit is provided without offset
	}

	if !limitExists && offsetExists {
		ep.Logger.InvalidInput("Missing limit parameter")
		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "offset parameter provided without limit")
		return nil, false
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		ep.Logger.InvalidInput("Failed to parse limit parameter", "limit", limitString, "error", err)
		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Invalid limit parameter: "+err.Error())
		return nil, false
	}
	if limit < 1 {
		ep.Logger.InvalidRequest("Invalid limit parameter", "limit", limit)
		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Limit must be greater than 0")
		return nil, false
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		ep.Logger.InvalidInput("Failed to parse offset parameter", "offset", offsetString, "error", err)
		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Invalid offset parameter: "+err.Error())
		return nil, false
	}
	if offset < 0 {
		ep.Logger.InvalidRequest("Invalid offset parameter", "offset", offset)
		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Offset must be 0 or greater")
		return nil, false
	}

	rowSelection := struct {
		limit  int
		offset int
	}{
		limit:  limit,
		offset: offset,
	}

	return &rowSelection, true
}

func (ep *getSalesEndpoint) parseOrder() (bool, bool) {
	if order, exists := ep.Context.GetQuery("order"); exists {
		if order != "antichronological" {
			ep.Logger.InvalidInput("Invalid order parameter", "order", order)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Order must be 'antichronological'")
			return false, false
		}
		return true, true
	}

	return false, true
}
