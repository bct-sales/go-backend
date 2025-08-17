package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	rest "bctbackend/server/shared"
	"errors"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type ListCashierSaleData struct {
	SaleId            models.Id           `json:"saleId"`
	TransactionTime   rest.DateTime       `json:"transactionTime"`
	ItemCount         int                 `json:"itemCount"`
	TotalPriceInCents models.MoneyInCents `json:"totalPriceInCents"`
}

type ListCashierSalesSuccessResponse struct {
	Sales []*ListCashierSaleData `json:"sales"`
}

type listCashierSalesEndpoint struct {
	context *gin.Context
	db      Database
	userId  models.Id
	roleId  models.RoleId
	logger  logger.Logger
}

func ListCashierSales(arguments *HandlerFunctionArguments) {
	endpoint := &listCashierSalesEndpoint{
		context: arguments.Context,
		db:      arguments.Database,
		userId:  arguments.UserId,
		roleId:  arguments.RoleId,
		logger:  arguments.Logger,
	}

	endpoint.Execute()
}

func (endpoint *listCashierSalesEndpoint) Execute() {
	uriCashierId, ok := endpoint.extractCashierIdFromUri()
	if !ok {
		return
	}

	var saleSummaries []*models.SaleSummary
	if err := queries.GetCashierSales(endpoint.db, uriCashierId, queries.CollectTo(&saleSummaries)); err != nil {
		endpoint.logger.InternalError("Failed to retrieve cashier sales for user %d: %v", uriCashierId, err)
		failure_response.Unknown(endpoint.context, "Could not retrieve cashier sales: "+err.Error())
		return
	}

	successResponse := ListCashierSalesSuccessResponse{
		Sales: algorithms.Map(saleSummaries, func(saleSummary *models.SaleSummary) *ListCashierSaleData {
			return endpoint.convertSaleSummaryToData(saleSummary)
		}),
	}

	endpoint.context.IndentedJSON(http.StatusOK, successResponse)
}

func (endpoint *listCashierSalesEndpoint) convertSaleSummaryToData(saleSummary *models.SaleSummary) *ListCashierSaleData {
	return &ListCashierSaleData{
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
func (endpoint *listCashierSalesEndpoint) extractCashierIdFromUri() (models.Id, bool) {
	var uriParameters struct {
		CashierId string `binding:"required" uri:"id"`
	}
	if err := endpoint.context.ShouldBindUri(&uriParameters); err != nil {
		endpoint.logger.InvalidInput("Failed to bind URI parameters: %v", err)
		failure_response.InvalidUriParameters(endpoint.context, err.Error())
		return 0, false
	}

	uriUserId, err := models.ParseId(uriParameters.CashierId)
	if err != nil {
		endpoint.logger.InvalidInput("Failed to parse cashier ID \"%s\" from URI: %v", uriParameters.CashierId, err)
		failure_response.InvalidUserId(endpoint.context, err.Error())
		return 0, false
	}

	if !endpoint.ensureUserHasPermission(uriUserId) {
		return 0, false
	}

	return uriUserId, true
}

func (endpoint *listCashierSalesEndpoint) ensureUserHasPermission(queriedUser models.Id) bool {
	user, err := queries.GetUserWithId(endpoint.db, endpoint.userId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should not happen, as the userId is from the logged-in user
			endpoint.logger.Bug("Logged in user does not exist")
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
			endpoint.logger.InvalidRequest("User tried to access sales of cashier %d, but is not the owning cashier", queriedUser)
			failure_response.Forbidden(endpoint.context, "wrong_role", "Only accessible to owning cashiers or admins")
			return false
		}

		return true
	}

	endpoint.logger.InvalidRequest("User tried to access sales of cashier %d, but is not a cashier or admin", queriedUser)
	failure_response.Forbidden(endpoint.context, "wrong_role", "Only accessible to owning cashiers or admins")
	return false
}
