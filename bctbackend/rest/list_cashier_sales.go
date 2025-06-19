package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"errors"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetCashierSaleData struct {
	SaleId            models.Id           `json:"saleId"`
	TransactionTime   rest.DateTime       `json:"transactionTime"`
	ItemCount         int                 `json:"itemCount"`
	TotalPriceInCents models.MoneyInCents `json:"totalPriceInCents"`
}

type GetCashierSalesSuccessResponse struct {
	Sales []*GetCashierSaleData `json:"sales"`
}

type getCashierSalesEndpoint struct {
	context *gin.Context
	db      *sql.DB
	userId  models.Id
	roleId  models.RoleId
}

func GetCashierSales(context *gin.Context, db *sql.DB, userId models.Id, roleId models.RoleId) {
	endpoint := &getCashierSalesEndpoint{
		context: context,
		db:      db,
		userId:  userId,
		roleId:  roleId,
	}

	endpoint.Execute()
}

func (ep *getCashierSalesEndpoint) Execute() {
	if !ep.ensureUserIsCashier() {
		return
	}

	uriCashierId, ok := ep.extractCashierIdFromUri()
	if !ok {
		return
	}

	var saleSummaries []*models.SaleSummary
	if err := queries.GetCashierSales(ep.db, uriCashierId, queries.CollectTo(&saleSummaries)); err != nil {
		failure_response.Unknown(ep.context, "Could not retrieve cashier sales: "+err.Error())
		return
	}

	successResponse := GetCashierSalesSuccessResponse{
		Sales: algorithms.Map(saleSummaries, func(saleSummary *models.SaleSummary) *GetCashierSaleData {
			return ep.convertSaleSummaryToData(saleSummary)
		}),
	}

	ep.context.IndentedJSON(http.StatusOK, successResponse)
}

func (ep *getCashierSalesEndpoint) convertSaleSummaryToData(saleSummary *models.SaleSummary) *GetCashierSaleData {
	return &GetCashierSaleData{
		SaleId:            saleSummary.SaleID,
		TransactionTime:   rest.ConvertTimestampToDateTime(saleSummary.TransactionTime),
		ItemCount:         saleSummary.ItemCount,
		TotalPriceInCents: saleSummary.TotalPriceInCents,
	}
}

func (ep *getCashierSalesEndpoint) ensureUserIsCashier() bool {
	if !ep.roleId.IsCashier() {
		failure_response.Forbidden(ep.context, "wrong_role", "Only accessible to cashiers")
		return false
	}
	return true
}

func (endpoint *getCashierSalesEndpoint) extractCashierIdFromUri() (models.Id, bool) {
	var uriParameters struct {
		CashierId string `uri:"id" binding:"required"`
	}
	if err := endpoint.context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(endpoint.context, err.Error())
		return 0, false
	}

	uriUserId, err := models.ParseId(uriParameters.CashierId)
	if err != nil {
		failure_response.InvalidUserId(endpoint.context, err.Error())
		return 0, false
	}

	if err := queries.EnsureUserExistsAndHasRole(endpoint.db, uriUserId, models.NewCashierRoleId()); err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			failure_response.UnknownUser(endpoint.context, err.Error())
			return 0, false
		}

		if errors.Is(err, dberr.ErrWrongRole) {
			failure_response.WrongUser(endpoint.context, "Can only list sales for cashiers")
			return 0, false
		}

		failure_response.Unknown(endpoint.context, "Could not check user role: "+err.Error())
		return 0, false
	}

	if endpoint.userId != uriUserId {
		failure_response.WrongSeller(endpoint.context, "Logged in user does not match URI cashier ID")
		return 0, false
	}

	return uriUserId, true
}
