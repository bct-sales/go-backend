package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"github.com/gin-gonic/gin"
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
	rowSelection               *queries.RowRange
	orderedAntiChronologically bool
}

func GetSales(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := &getSalesEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context, arguments.Database)
}

func (ep *getSalesEndpoint) execute(context *gin.Context, database Database) {
	if !ep.ensureUserIsAdmin(context) {
		return
	}

	queryParameters, queryParametersOk := ep.parseQueryParameters(context)
	if !queryParametersOk {
		return
	}

	response, responseOk := ep.fetchData(context, database, queryParameters)
	if !responseOk {
		return
	}

	context.IndentedJSON(http.StatusOK, response)
}

func (ep *getSalesEndpoint) fetchData(context *gin.Context, database Database, queryParameters *getSalesQueryParameters) (*ListSalesSuccessResponse, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	transaction, err := database.StartTransaction(ctx)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to create transaction")

		failure_response.Unknown(context, "Failed to create transaction: "+err.Error())

		return nil, false
	}
	defer transaction.RollbackIfNotCommitted()

	sales, getSalesOk := ep.getSales(context, transaction, queryParameters)
	if !getSalesOk {
		return nil, false
	}

	saleCount, countSalesOk := ep.countSales(context, transaction)
	if !countSalesOk {
		return nil, false
	}

	totalSaleValue, getTotalSalesValueOk := ep.getTotalSalesValue(context, transaction)
	if !getTotalSalesValueOk {
		return nil, false
	}

	itemStatistics := ep.getItemStatistics(context, transaction)
	if itemStatistics == nil {
		return nil, false
	}

	distinctSoldItemCount, totalSoldItemCount, countSoldItemsOk := ep.countSoldItems(context, transaction)
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
		logger.AddInformation("error", err)
		logger.InternalError("Failed to commit transaction")

		failure_response.Unknown(context, "Failed to commit transaction: "+err.Error())

		return nil, false
	}

	return &response, true
}

func (ep *getSalesEndpoint) getItemStatistics(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier) *queries.ItemStatisticsResult {
	logger := ep.Logger
	ctx := context.Request.Context()
	query := queries.NewGetItemStatisticsQuery()
	query.WithHidden(false)
	itemCountResult, err := query.Execute(ctx, transaction)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get item count")

		failure_response.Unknown(context, "Failed to get item count: "+err.Error())

		return nil
	}

	return itemCountResult
}

func (ep *getSalesEndpoint) countSoldItems(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier) (int, int, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	counts, err := queries.CountSoldItems(ctx, transaction)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get sold item count")

		failure_response.Unknown(context, "Failed to get sold item count: "+err.Error())

		return 0, 0, false
	}

	return counts.Distinct, counts.IncludeMultiples, true
}

func (ep *getSalesEndpoint) countSales(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier) (int, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	saleCount, err := queries.CountSales(ctx, transaction)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get sales count")

		failure_response.Unknown(context, "Failed to get sales count: "+err.Error())

		return 0, false
	}

	return saleCount, true
}

func (ep *getSalesEndpoint) getTotalSalesValue(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier) (models.MoneyInCents, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	totalValue, err := queries.GetTotalSalesValue(ctx, transaction)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get total sales value")

		failure_response.Unknown(context, "Failed to get total sales value: "+err.Error())

		return 0, false
	}

	return totalValue, true

}

func (ep *getSalesEndpoint) ensureUserIsAdmin(context *gin.Context) bool {
	logger := ep.Logger

	if ep.RoleID != models.NewAdminRoleID() {
		logger.InvalidRequest("Unauthorized access to list all sales")

		failure_response.WrongRole(context, "Only admins can list all items")

		return false
	}

	return true
}

func (ep *getSalesEndpoint) getSales(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier, queryParameters *getSalesQueryParameters) ([]*ListSalesSaleData, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
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

	if err := query.Execute(ctx, transaction, processSale); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get sales")

		failure_response.Unknown(context, "Failed to get sales: "+err.Error())

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

		query.WithRowRange(limit, offset)
	}

	if queryParameters.orderedAntiChronologically {
		query.OrderedAntiChronologically()
	}

	return query
}

func (ep *getSalesEndpoint) parseQueryParameters(context *gin.Context) (*getSalesQueryParameters, bool) {
	startID, ok := ep.parseStartID(context)
	if !ok {
		return nil, false
	}

	rowSelection := ep.parseRowRangeQueryParameters(context)
	if rowSelection == nil {
		return nil, false
	}

	order, ok := ep.parseOrderQueryParameter(context)
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

func (ep *getSalesEndpoint) parseStartID(context *gin.Context) (*models.ID, bool) {
	logger := ep.Logger
	if startIDStr, exists := context.GetQuery("startId"); exists {
		startID, err := models.ParseID(startIDStr)
		if err != nil {
			logger.AddInformation("startId", startIDStr)
			logger.AddInformation("error", err)
			logger.InvalidInput("Failed to parse startId parameter")

			failure_response.BadRequest(context, "invalid_uri_parameters", "Invalid startId parameter: "+err.Error())

			return nil, false
		}
		return &startID, true
	}

	return nil, true
}
