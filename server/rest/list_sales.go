package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"net/http"
)

type ListSalesSaleData struct {
	SaleID            models.ID           `binding:"required" json:"saleId"`
	CashierID         models.ID           `binding:"required" json:"cashierId"`
	TransactionTime   rest.DateTime       `binding:"required" json:"transactionTime"`
	ItemCount         int                 `binding:"required" json:"itemCount"`
	TotalPriceInCents models.MoneyInCents `binding:"required" json:"totalPriceInCents"`
}

type ListSalesSuccessResponse struct {
	Sales                 []*ListSalesSaleData `json:"sales"`
	ItemCount             int                  `json:"itemCount"`
	TotalItemValue        models.MoneyInCents  `json:"totalItemValue"`
	DistinctSoldItemCount int                  `json:"distinctSoldItemCount"`
	TotalSoldItemCount    int                  `json:"totalSoldItemCount"`
	SaleCount             int                  `json:"saleCount"`
	TotalSaleValue        models.MoneyInCents  `json:"totalSaleValueInCents"`
}

type getSalesEndpoint struct {
	Endpoint
}

type getSalesQueryParameters struct {
	startID                    *models.ID
	rowSelection               *queries.RowSelection
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
	defer transaction.RollbackIfNotCommitted()

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

	itemStatistics := ep.getItemStatistics(transaction)
	if itemStatistics == nil {
		return nil, false
	}

	distinctSoldItemCount, totalSoldItemCount, countSoldItemsOk := ep.countSoldItems(transaction)
	if !countSoldItemsOk {
		return nil, false
	}

	response := ListSalesSuccessResponse{
		Sales:                 sales,
		ItemCount:             itemStatistics.ItemCount,
		TotalItemValue:        itemStatistics.TotalValueInCents,
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

func (ep *getSalesEndpoint) getItemStatistics(transaction *queries.TransactionalDatabaseQuerier) *queries.ItemStatisticsResult {
	itemCountResult, err := queries.GetItemStatistics(transaction, queries.OnlyVisibleItems)

	if err != nil {
		ep.Logger.InternalError("Failed to get item count", "error", err)
		failure_response.Unknown(ep.Context, "Failed to get item count: "+err.Error())
		return nil
	}

	return itemCountResult
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
	if ep.RoleID != models.NewAdminRoleID() {
		ep.Logger.InvalidRequest("Unauthorized access to list all sales", "userID", ep.UserID, "roleID", ep.RoleID)
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

	if queryParameters.startID != nil {
		query.WithIDGreaterThanOrEqualTo(*queryParameters.startID)
	}

	if queryParameters.rowSelection != nil {
		var limit uint64
		var offset uint64

		if queryParameters.rowSelection.Limit != nil {
			limit = *queryParameters.rowSelection.Limit
		} else {
			limit = 10000
		}

		if queryParameters.rowSelection.Offset != nil {
			offset = *queryParameters.rowSelection.Offset
		} else {
			offset = 0
		}

		query.WithRowSelection(limit, offset)
	}

	if queryParameters.orderedAntiChronologically {
		query.OrderedAntiChronologically()
	}

	return query
}

func (ep *getSalesEndpoint) parseQueryParameters() (*getSalesQueryParameters, bool) {
	startID, ok := ep.parseStartID()
	if !ok {
		return nil, false
	}

	rowSelection := ep.parseRowSelectionQueryParameters()
	if rowSelection == nil {
		return nil, false
	}

	order, ok := ep.parseOrderQueryParameter()
	if !ok {
		return nil, false
	}
	antiChronologicalOrder := order == queries.OrderAntiChronological

	queryParameters := getSalesQueryParameters{
		startID:                    startID,
		rowSelection:               rowSelection,
		orderedAntiChronologically: antiChronologicalOrder,
	}

	return &queryParameters, true
}

func (ep *getSalesEndpoint) parseStartID() (*models.ID, bool) {
	if startIDStr, exists := ep.Context.GetQuery("startId"); exists {
		startID, err := models.ParseID(startIDStr)
		if err != nil {
			ep.Logger.InvalidInput("Failed to parse startId parameter", "startId", startIDStr, "error", err)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Invalid startId parameter: "+err.Error())
			return nil, false
		}
		return &startID, true
	}

	return nil, true
}
