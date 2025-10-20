package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
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

func ListCashierSales(arguments *HandlerFunctionArguments) {
	endpoint := &listCashierSalesEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type listCashierSalesEndpoint struct {
	Endpoint
}

func (ep *listCashierSalesEndpoint) execute() {
	uriCashierID, ok := ep.extractCashierIDFromUri()
	if !ok {
		return
	}

	rowSelection := ep.parseRowRangeQueryParameters()
	if rowSelection == nil {
		return
	}

	order, orderOk := ep.parseOrderQueryParameter()
	if !orderOk {
		return
	}

	saleSummaries, saleSummariesOk := ep.getSaleSummariesFromDatabase(uriCashierID, order, rowSelection)
	if !saleSummariesOk {
		return
	}

	saleCount, saleCountOk := ep.getCashierSaleCount(uriCashierID)
	if !saleCountOk {
		return
	}

	ep.sendSuccessResponse(saleCount, saleSummaries)
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
func (ep *listCashierSalesEndpoint) extractCashierIDFromUri() (models.ID, bool) {
	logger := ep.Logger
	var uriParameters struct {
		CashierID string `binding:"required" uri:"id"`
	}

	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to bind URI parameters")

		failure_response.InvalidUriParameters(ep.Context, err.Error())

		return 0, false
	}

	uriUserID, err := models.ParseID(uriParameters.CashierID)
	if err != nil {
		logger.AddInformation("cashierID", uriParameters.CashierID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse cashier ID")

		failure_response.InvalidUserID(ep.Context, err.Error())

		return 0, false
	}

	if !ep.ensureUserHasPermission(uriUserID) {
		return 0, false
	}

	return uriUserID, true
}

func (ep *listCashierSalesEndpoint) ensureUserHasPermission(queriedUser models.ID) bool {
	logger := ep.Logger
	user, err := queries.GetUserWithID(ep.Database, ep.UserID)

	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should not happen, as the userID is from the logged-in user
			logger.Bug("Logged in user does not exist")

			failure_response.Unknown(ep.Context, "Bug: logged in user does not exist")

			return false
		}

		logger.InternalError("Could not find logged in user in database")

		failure_response.Unknown(ep.Context, "Could not retrieve logged in user: "+err.Error())

		return false
	}

	if user.RoleID.IsAdmin() {
		return true
	}

	if user.RoleID.IsCashier() {
		loggedInUser := ep.UserID

		if loggedInUser != queriedUser {
			logger.InvalidRequest("User tried to access sales of other cashier")

			failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to owning cashiers or admins")

			return false
		}

		return true
	}

	logger.InvalidRequest("Non-cashier tried to access sales")

	failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to owning cashiers or admins")

	return false
}

func (ep *listCashierSalesEndpoint) getSaleSummariesFromDatabase(uriCashierID models.ID, order queries.Order, rowSelection *queries.RowRange) ([]*models.SaleSummary, bool) {
	logger := ep.Logger
	var saleSummaries []*models.SaleSummary

	if err := queries.GetCashierSales(ep.Database, uriCashierID, queries.CollectTo(&saleSummaries), order, rowSelection); err != nil {
		logger.InternalError("Failed to retrieve cashier sales for user")

		failure_response.Unknown(ep.Context, "Could not retrieve cashier sales: "+err.Error())

		return nil, false
	}

	return saleSummaries, true
}

func (ep *listCashierSalesEndpoint) sendSuccessResponse(saleCount int, saleSummaries []*models.SaleSummary) {
	successResponse := ListCashierSalesSuccessResponse{
		Sales: algorithms.Map(saleSummaries, func(saleSummary *models.SaleSummary) *ListCashierSaleData {
			return ep.convertSaleSummaryToData(saleSummary)
		}),
		SaleCount: saleCount,
	}
	ep.Context.IndentedJSON(http.StatusOK, successResponse)
}

func (ep *listCashierSalesEndpoint) getCashierSaleCount(uriCashierID models.ID) (int, bool) {
	logger := ep.Logger
	count, err := queries.CountCashierSales(ep.Database, uriCashierID)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to count cashier sales")

		failure_response.Unknown(ep.Context, "Could not count cashier sales: "+err.Error())

		return 0, false
	}

	return count, true
}
