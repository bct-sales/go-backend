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
	SaleId            models.Id           `json:"saleId"`
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
	uriCashierId, ok := ep.extractCashierIdFromUri()
	if !ok {
		return
	}

	rowSelection := ep.parseRowSelectionQueryParameters()
	if rowSelection == nil {
		return
	}

	order, orderOk := ep.parseOrderQueryParameter()
	if !orderOk {
		return
	}

	saleSummaries, saleSummariesOk := ep.getSaleSummariesFromDatabase(uriCashierId, order, rowSelection)
	if !saleSummariesOk {
		return
	}

	saleCount, saleCountOk := ep.getCashierSaleCount(uriCashierId)
	if !saleCountOk {
		return
	}

	ep.sendSuccessResponse(saleCount, saleSummaries)
}

func (ep *listCashierSalesEndpoint) convertSaleSummaryToData(saleSummary *models.SaleSummary) *ListCashierSaleData {
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
func (ep *listCashierSalesEndpoint) extractCashierIdFromUri() (models.Id, bool) {
	var uriParameters struct {
		CashierId string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		ep.Logger.InvalidInput("Failed to bind URI parameters: %v", err)
		failure_response.InvalidUriParameters(ep.Context, err.Error())
		return 0, false
	}

	uriUserId, err := models.ParseId(uriParameters.CashierId)
	if err != nil {
		ep.Logger.InvalidInput("Failed to parse cashier ID \"%s\" from URI: %v", uriParameters.CashierId, err)
		failure_response.InvalidUserId(ep.Context, err.Error())
		return 0, false
	}

	if !ep.ensureUserHasPermission(uriUserId) {
		return 0, false
	}

	return uriUserId, true
}

func (ep *listCashierSalesEndpoint) ensureUserHasPermission(queriedUser models.Id) bool {
	user, err := queries.GetUserWithId(ep.Database, ep.UserId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should not happen, as the userId is from the logged-in user
			ep.Logger.Bug("Logged in user does not exist")
			failure_response.Unknown(ep.Context, "Bug: logged in user does not exist")
			return false
		}
		failure_response.Unknown(ep.Context, "Could not retrieve logged in user: "+err.Error())
		return false
	}

	if user.RoleId.IsAdmin() {
		return true
	}

	if user.RoleId.IsCashier() {
		loggedInUser := ep.UserId

		if loggedInUser != queriedUser {
			ep.Logger.InvalidRequest("User tried to access sales of cashier %d, but is not the owning cashier", queriedUser)
			failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to owning cashiers or admins")
			return false
		}

		return true
	}

	ep.Logger.InvalidRequest("User tried to access sales of cashier %d, but is not a cashier or admin", queriedUser)
	failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to owning cashiers or admins")
	return false
}

func (ep *listCashierSalesEndpoint) getSaleSummariesFromDatabase(uriCashierId models.Id, order queries.Order, rowSelection *queries.RowSelection) ([]*models.SaleSummary, bool) {
	var saleSummaries []*models.SaleSummary
	if err := queries.GetCashierSales(ep.Database, uriCashierId, queries.CollectTo(&saleSummaries), order, rowSelection); err != nil {
		ep.Logger.InternalError("Failed to retrieve cashier sales for user %d: %v", uriCashierId, err)
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

func (ep *listCashierSalesEndpoint) getCashierSaleCount(uriCashierId models.Id) (int, bool) {
	count, err := queries.CountCashierSales(ep.Database, uriCashierId)
	if err != nil {
		ep.Logger.InternalError("Failed to count cashier sales for user %d: %v", uriCashierId, err)
		failure_response.Unknown(ep.Context, "Could not count cashier sales: "+err.Error())
		return 0, false
	}

	return count, true
}
