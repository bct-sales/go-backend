package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ListCashierSaleData struct {
	SaleID            models.ID           `json:"saleId"`
	TransactionTime   rest.DateTime       `json:"transactionTime"`
	ItemCount         int                 `json:"itemCount"`
	TotalPriceInCents models.MoneyInCents `json:"totalPriceInCents"`
}

type ListCashierSalesSuccessResponse struct {
	Sales     []*ListCashierSaleData `json:"sales"`
	SaleCount int                    `json:"saleCount"`
}

func ListCashierSales(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := &listCashierSalesEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type listCashierSalesEndpoint struct {
	Endpoint
}

func (ep *listCashierSalesEndpoint) execute(context *gin.Context) {
	uriCashierID, ok := ep.extractCashierIDFromUri(context)
	if !ok {
		return
	}

	rowSelection := ep.parseRowRangeQueryParameters(context)
	if rowSelection == nil {
		return
	}

	order, orderOk := ep.parseOrderQueryParameter(context)
	if !orderOk {
		return
	}

	saleSummaries, saleSummariesOk := ep.getSaleSummariesFromDatabase(context, uriCashierID, order, rowSelection)
	if !saleSummariesOk {
		return
	}

	saleCount, saleCountOk := ep.getCashierSaleCount(context, uriCashierID)
	if !saleCountOk {
		return
	}

	ep.sendSuccessResponse(context, saleCount, saleSummaries)
}

func (ep *listCashierSalesEndpoint) convertSaleSummaryToData(saleSummary *models.SaleSummary) *ListCashierSaleData {
	return &ListCashierSaleData{
		SaleID:            saleSummary.SaleID,
		TransactionTime:   rest.ConvertTimestampToDateTime(saleSummary.TransactionTime),
		ItemCount:         saleSummary.ItemCount,
		TotalPriceInCents: saleSummary.TotalPriceInCents,
	}
}

// extractCashierIDFromUri extracts the cashier ID from the URI and validates it.
// It returns the cashier ID and a boolean indicating success or failure.
// If the extraction or validation fails, it sends an appropriate error response.
// False indicates failure, true indicates success.
func (ep *listCashierSalesEndpoint) extractCashierIDFromUri(context *gin.Context) (models.ID, bool) {
	logger := ep.Logger
	var uriParameters struct {
		CashierID string `binding:"required" uri:"id"`
	}

	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to bind URI parameters")

		failure_response.InvalidUriParameters(context, err.Error())

		return 0, false
	}

	uriUserID, err := models.ParseID(uriParameters.CashierID)
	if err != nil {
		logger.AddInformation("cashierID", uriParameters.CashierID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse cashier ID")

		failure_response.InvalidUserID(context, err.Error())

		return 0, false
	}

	if !ep.ensureUserHasPermission(context, uriUserID) {
		return 0, false
	}

	return uriUserID, true
}

func (ep *listCashierSalesEndpoint) ensureUserHasPermission(context *gin.Context, queriedUser models.ID) bool {
	logger := ep.Logger
	ctx := context.Request.Context()
	user, err := queries.GetUserWithID(ctx, ep.Database, ep.UserID)

	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should not happen, as the userID is from the logged-in user
			logger.Bug("Logged in user does not exist")

			failure_response.Unknown(context, "Bug: logged in user does not exist")

			return false
		}

		logger.InternalError("Could not find logged in user in database")

		failure_response.Unknown(context, "Could not retrieve logged in user: "+err.Error())

		return false
	}

	if user.RoleID.IsAdmin() {
		return true
	}

	if user.RoleID.IsCashier() {
		loggedInUser := ep.UserID

		if loggedInUser != queriedUser {
			logger.InvalidRequest("User tried to access sales of other cashier")

			failure_response.Forbidden(context, "wrong_role", "Only accessible to owning cashiers or admins")

			return false
		}

		return true
	}

	logger.InvalidRequest("Non-cashier tried to access sales")

	failure_response.Forbidden(context, "wrong_role", "Only accessible to owning cashiers or admins")

	return false
}

func (ep *listCashierSalesEndpoint) getSaleSummariesFromDatabase(context *gin.Context, uriCashierID models.ID, order queries.Order, rowSelection *queries.RowRange) ([]*models.SaleSummary, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	var saleSummaries []*models.SaleSummary

	if err := queries.GetCashierSales(ctx, ep.Database, uriCashierID, queries.CollectTo(&saleSummaries), order, rowSelection); err != nil {
		logger.InternalError("Failed to retrieve cashier sales for user")

		failure_response.Unknown(context, "Could not retrieve cashier sales: "+err.Error())

		return nil, false
	}

	return saleSummaries, true
}

func (ep *listCashierSalesEndpoint) sendSuccessResponse(context *gin.Context, saleCount int, saleSummaries []*models.SaleSummary) {
	successResponse := ListCashierSalesSuccessResponse{
		Sales: algorithms.Map(saleSummaries, func(saleSummary *models.SaleSummary) *ListCashierSaleData {
			return ep.convertSaleSummaryToData(saleSummary)
		}),
		SaleCount: saleCount,
	}
	context.IndentedJSON(http.StatusOK, successResponse)
}

func (ep *listCashierSalesEndpoint) getCashierSaleCount(context *gin.Context, uriCashierID models.ID) (int, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	count, err := queries.CountCashierSales(ctx, ep.Database, uriCashierID)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to count cashier sales")

		failure_response.Unknown(context, "Could not count cashier sales: "+err.Error())

		return 0, false
	}

	return count, true
}
