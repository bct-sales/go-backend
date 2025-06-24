package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"database/sql"
	"net/http"
	"strconv"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type ListSalesSaleData struct {
	SaleID            models.Id           `json:"saleId"`
	CashierID         models.Id           `json:"cashierId"`
	TransactionTime   rest.DateTime       `json:"transactionTime"`
	ItemCount         int                 `json:"itemCount"`
	TotalPriceInCents models.MoneyInCents `json:"totalPriceInCents"`
}

type ListSalesSuccessResponse struct {
	Sales          []*ListSalesSaleData `json:"sales"`
	ItemCount      int                  `json:"itemCount"`
	SoldItemCount  int                  `json:"soldItemCount"`
	SaleCount      int                  `json:"saleCount"`
	TotalSaleValue models.MoneyInCents  `json:"totalSaleValueInCents"`
}

type getSalesEndpoint struct {
	context *gin.Context
	db      *sql.DB
	userId  models.Id
	roleId  models.RoleId
}

type getSalesQueryParameters struct {
	startId      *models.Id
	rowSelection *struct {
		limit  int
		offset int
	}
	orderedAntiChronologically bool
}

func GetSales(context *gin.Context, configuration *configuration.Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	endpoint := &getSalesEndpoint{
		context: context,
		db:      db,
		userId:  userId,
		roleId:  roleId,
	}

	endpoint.execute()
}

func (ep *getSalesEndpoint) execute() {
	if !ep.ensureUserIsAdmin() {
		return
	}

	queryParameters, ok := ep.parseQueryParameters()
	if !ok {
		return
	}

	sales, ok := ep.getSales(queryParameters)
	if !ok {
		return
	}

	saleCount, ok := ep.getSaleCount()
	if !ok {
		return
	}

	totalSaleValue, ok := ep.getTotalSalesValue()
	if !ok {
		return
	}

	itemCount, ok := ep.getItemCount()
	if !ok {
		return
	}

	soldItemCount, ok := ep.getSoldItemCount()
	if !ok {
		return
	}

	response := ListSalesSuccessResponse{
		Sales:          sales,
		SaleCount:      saleCount,
		TotalSaleValue: totalSaleValue,
		ItemCount:      itemCount,
		SoldItemCount:  soldItemCount,
	}

	ep.context.IndentedJSON(http.StatusOK, response)
}

func (ep *getSalesEndpoint) getItemCount() (int, bool) {
	soldItemCount, err := queries.CountItems(ep.db, queries.OnlyVisibleItems)
	if err != nil {
		slog.Error("Failed to get sold item count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sold item count: "+err.Error())
		return 0, false
	}
	return soldItemCount, true
}

func (ep *getSalesEndpoint) getSoldItemCount() (int, bool) {
	soldItemCount, err := queries.GetSoldItemsCount(ep.db)

	if err != nil {
		slog.Error("Failed to get sold item count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sold item count: "+err.Error())
		return 0, false
	}

	return soldItemCount, true
}

func (ep *getSalesEndpoint) getSaleCount() (int, bool) {
	saleCount, err := queries.GetSalesCount(ep.db)

	if err != nil {
		slog.Error("Failed to get sales count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sales count: "+err.Error())
		return 0, false
	}

	return saleCount, true
}

func (ep *getSalesEndpoint) getTotalSalesValue() (models.MoneyInCents, bool) {
	totalValue, err := queries.GetTotalSalesValue(ep.db)

	if err != nil {
		slog.Error("Failed to get total sales value", "error", err)
		failure_response.Unknown(ep.context, "Failed to get total sales value: "+err.Error())
		return 0, false
	}

	return totalValue, true

}

func (ep *getSalesEndpoint) ensureUserIsAdmin() bool {
	if ep.roleId != models.NewAdminRoleId() {
		slog.Error("Unauthorized access to list all sales", "userId", ep.userId, "roleId", ep.roleId)
		failure_response.WrongRole(ep.context, "Only admins can list all items")
		return false
	}

	return true
}

func (ep *getSalesEndpoint) getSales(queryParameters *getSalesQueryParameters) ([]*ListSalesSaleData, bool) {
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

	if err := query.Execute(ep.db, processSale); err != nil {
		slog.Error("Failed to get sales", "error", err)
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
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "offset parameter provided without limit")
		return nil, false
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Invalid limit parameter: "+err.Error())
		return nil, false
	}
	if limit < 1 {
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Limit must be greater than 0")
		return nil, false
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Invalid offset parameter: "+err.Error())
		return nil, false
	}
	if offset < 0 {
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
			failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Order must be 'antichronological'")
			return false, false
		}
		return true, true
	}
	return false, true
}
