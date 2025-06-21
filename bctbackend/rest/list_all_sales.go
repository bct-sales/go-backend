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
	Sales []*ListSalesSaleData `json:"sales"`
}

type getAllSalesEndpoint struct {
	context *gin.Context
	db      *sql.DB
	userId  models.Id
	roleId  models.RoleId
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

	if err := queries.NewGetSalesQuery().Execute(ep.db, processSale); err != nil {
		slog.Error("Failed to get sales", "error", err)
		failure_response.Unknown(ep.context, "Failed to get sales: "+err.Error())
		return
	}

	response := ListSalesSuccessResponse{
		Sales: sales,
	}

	ep.context.IndentedJSON(http.StatusOK, response)
}

func (ep *getAllSalesEndpoint) ensureUserIsAdmin() bool {
	if ep.roleId != models.NewAdminRoleId() {
		slog.Error("Unauthorized access to list all sales", "userId", ep.userId, "roleId", ep.roleId)
		failure_response.WrongRole(ep.context, "Only admins can list all items")
		return false
	}

	return true
}
