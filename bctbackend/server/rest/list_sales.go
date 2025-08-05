package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	rest "bctbackend/server/shared"
	"database/sql"
	"net/http"
	"strconv"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
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
	context *gin.Context
	userId  models.Id
	roleId  models.RoleId
	logger  logger.Logger
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
		context: arguments.Context,
		userId:  arguments.UserId,
		roleId:  arguments.RoleId,
		logger:  arguments.Logger,
	}

	endpoint.execute(arguments.Database)
}

func (ep *getSalesEndpoint) execute(database *sql.DB) {
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

	ep.context.IndentedJSON(http.StatusOK, response)
}

func (ep *getSalesEndpoint) fetchData(database *sql.DB, queryParameters *getSalesQueryParameters) (*ListSalesSuccessResponse, bool) {
	transaction, err := queries.NewTransaction(database)
	if err != nil {
		ep.logger.InternalError("Failed to create transaction", err)
		failure_response.Unknown(ep.context, "Failed to create transaction: "+err.Error())
		return nil, false
	}
	defer transaction.Commit()

	sales, ok := ep.getSales(transaction, queryParameters)
	if !ok {
		transaction.Rollback()
		return nil, false
	}

	saleCount, ok := ep.countSales(transaction)
	if !ok {
		transaction.Rollback()
		return nil, false
	}

	totalSaleValue, ok := ep.getTotalSalesValue(transaction)
	if !ok {
		transaction.Rollback()
		return nil, false
	}

	itemCount, ok := ep.countItems(transaction)
	if !ok {
		transaction.Rollback()
		return nil, false
	}

	distinctSoldItemCount, totalSoldItemCount, ok := ep.countSoldItems(transaction)
	if !ok {
		transaction.Rollback()
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

	return &response, true
}

func (ep *getSalesEndpoint) countItems(transaction *queries.Transaction) (int, bool) {
	soldItemCount, err := queries.CountItems(transaction, queries.OnlyVisibleItems)
	if err != nil {
		ep.logger.InternalError("Failed to get sold item count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sold item count: "+err.Error())
		return 0, false
	}
	return soldItemCount, true
}

func (ep *getSalesEndpoint) countSoldItems(transaction *queries.Transaction) (int, int, bool) {
	counts, err := queries.CountSoldItems(transaction)

	if err != nil {
		ep.logger.InternalError("Failed to get sold item count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sold item count: "+err.Error())
		return 0, 0, false
	}

	return counts.Distinct, counts.IncludeMultiples, true
}

func (ep *getSalesEndpoint) countSales(transaction *queries.Transaction) (int, bool) {
	saleCount, err := queries.CountSales(transaction)

	if err != nil {
		ep.logger.InternalError("Failed to get sales count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sales count: "+err.Error())
		return 0, false
	}

	return saleCount, true
}

func (ep *getSalesEndpoint) getTotalSalesValue(transaction *queries.Transaction) (models.MoneyInCents, bool) {
	totalValue, err := queries.GetTotalSalesValue(transaction)

	if err != nil {
		ep.logger.InternalError("Failed to get total sales value", "error", err)
		failure_response.Unknown(ep.context, "Failed to get total sales value: "+err.Error())
		return 0, false
	}

	return totalValue, true

}

func (ep *getSalesEndpoint) ensureUserIsAdmin() bool {
	if ep.roleId != models.NewAdminRoleId() {
		ep.logger.InvalidRequest("Unauthorized access to list all sales", "userId", ep.userId, "roleId", ep.roleId)
		failure_response.WrongRole(ep.context, "Only admins can list all items")
		return false
	}

	return true
}

func (ep *getSalesEndpoint) getSales(transaction *queries.Transaction, queryParameters *getSalesQueryParameters) ([]*ListSalesSaleData, bool) {
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
		ep.logger.InternalError("Failed to get sales", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sales: "+err.Error())
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
	if startIdStr, exists := ep.context.GetQuery("startId"); exists {
		startId, err := models.ParseId(startIdStr)
		if err != nil {
			ep.logger.InvalidInput("Failed to parse startId parameter", "startId", startIdStr, "error", err)
			failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Invalid startId parameter: "+err.Error())
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

	limitString, limitExists := ep.context.GetQuery("limit")
	offsetString, offsetExists := ep.context.GetQuery("offset")

	if !limitExists && !offsetExists {
		return nil, true
	}

	if limitExists && !offsetExists {
		offsetString = "0" // Default offset to 0 if limit is provided without offset
	}

	if !limitExists && offsetExists {
		ep.logger.InvalidInput("Missing limit parameter")
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "offset parameter provided without limit")
		return nil, false
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		ep.logger.InvalidInput("Failed to parse limit parameter", "limit", limitString, "error", err)
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Invalid limit parameter: "+err.Error())
		return nil, false
	}
	if limit < 1 {
		ep.logger.InvalidRequest("Invalid limit parameter", "limit", limit)
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Limit must be greater than 0")
		return nil, false
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		ep.logger.InvalidInput("Failed to parse offset parameter", "offset", offsetString, "error", err)
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Invalid offset parameter: "+err.Error())
		return nil, false
	}
	if offset < 0 {
		ep.logger.InvalidRequest("Invalid offset parameter", "offset", offset)
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Offset must be 0 or greater")
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
	if order, exists := ep.context.GetQuery("order"); exists {
		if order != "antichronological" {
			ep.logger.InvalidInput("Invalid order parameter", "order", order)
			failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Order must be 'antichronological'")
			return false, false
		}
		return true, true
	}

	return false, true
}
