package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
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

func GetCashierSales(context *gin.Context, configuration *configuration.Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	endpoint := &getCashierSalesEndpoint{
		context: context,
		db:      db,
		userId:  userId,
		roleId:  roleId,
	}

	endpoint.Execute()
}

func (ep *getCashierSalesEndpoint) Execute() {
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

// extractCashierIdFromUri extracts the cashier ID from the URI and validates it.
// It returns the cashier ID and a boolean indicating success or failure.
// If the extraction or validation fails, it sends an appropriate error response.
// False indicates failure, true indicates success.
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

	if !endpoint.ensureUserHasPermission(uriUserId) {
		return 0, false
	}

	return uriUserId, true
}

func (endpoint *getCashierSalesEndpoint) ensureUserHasPermission(queriedUser models.Id) bool {
	user, err := queries.GetUserWithId(endpoint.db, endpoint.userId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should not happen, as the userId is from the logged-in user
			failure_response.Unknown(endpoint.context, "Bug: logged in user does not exist")
			return false
		}
		failure_response.Unknown(endpoint.context, "Could not retrieve logged in user: "+err.Error())
		return false
	}

	if user.RoleId.IsAdmin() {
		return true
	}

	if user.RoleId.IsCashier() {
		loggedInUser := endpoint.userId

		if loggedInUser != queriedUser {
			failure_response.Forbidden(endpoint.context, "wrong_role", "Only accessible to owning cashiers or admins")
			return false
		}

		return true
	}

	failure_response.Forbidden(endpoint.context, "wrong_role", "Only accessible to owning cashiers or admins")
	return false
}
