package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"net/http"

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
	SaleCount      int                  `json:"saleCount"`
	TotalSaleValue models.MoneyInCents  `json:"totalSaleValue"`
}

type getAllSalesEndpoint struct {
	context *gin.Context
	db      *sql.DB
	userId  models.Id
	roleId  models.RoleId
}

type getAllSalesQueryParameters struct {
	startId *models.Id
}

func GetAllSales(context *gin.Context, configuration *Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	endpoint := &getAllSalesEndpoint{
		context: context,
		db:      db,
		userId:  userId,
		roleId:  roleId,
	}

	endpoint.execute()
}

func (ep *getAllSalesEndpoint) execute() {
	if !ep.ensureUserIsAdmin() {
		return
	}

	queryParameters, ok := ep.parseQueryParameters()
	if !ok {
		return
	}

	sales, ok := ep.getAllSales(queryParameters)
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

	response := ListSalesSuccessResponse{
		Sales:          sales,
		SaleCount:      saleCount,
		TotalSaleValue: totalSaleValue,
	}

	ep.context.IndentedJSON(http.StatusOK, response)
}

func (ep *getAllSalesEndpoint) getSaleCount() (int, bool) {
	saleCount, err := queries.GetSalesCount(ep.db)

	if err != nil {
		slog.Error("Failed to get sales count", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sales count: "+err.Error())
		return 0, false
	}

	return saleCount, true
}

func (ep *getAllSalesEndpoint) getTotalSalesValue() (models.MoneyInCents, bool) {
	totalValue, err := queries.GetTotalSalesValue(ep.db)

	if err != nil {
		slog.Error("Failed to get total sales value", "error", err)
		failure_response.Unknown(ep.context, "Failed to get total sales value: "+err.Error())
		return 0, false
	}

	return totalValue, true

}

func (ep *getAllSalesEndpoint) ensureUserIsAdmin() bool {
	if ep.roleId != models.NewAdminRoleId() {
		slog.Error("Unauthorized access to list all sales", "userId", ep.userId, "roleId", ep.roleId)
		failure_response.WrongRole(ep.context, "Only admins can list all items")
		return false
	}

	return true
}

func (ep *getAllSalesEndpoint) getAllSales(queryParameters *getAllSalesQueryParameters) ([]*ListSalesSaleData, bool) {
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

	query := queries.NewGetSalesQuery()

	if queryParameters.startId != nil {
		query.WithIdGreaterThanOrEqualTo(*queryParameters.startId)
	}

	if err := query.Execute(ep.db, processSale); err != nil {
		slog.Error("Failed to get sales", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sales: "+err.Error())
		return nil, false
	}

	return sales, true
}

func (ep *getAllSalesEndpoint) parseQueryParameters() (*getAllSalesQueryParameters, bool) {
	queryParameters := getAllSalesQueryParameters{
		startId: nil,
	}

	if startIdStr, exists := ep.context.GetQuery("startId"); exists {
		startId, err := models.ParseId(startIdStr)
		if err != nil {
			failure_response.BadRequest(ep.context, "invalid_uri_parameters", "Invalid startId parameter: "+err.Error())
			return nil, false
		}
		queryParameters.startId = &startId
	}

	return &queryParameters, true
}
