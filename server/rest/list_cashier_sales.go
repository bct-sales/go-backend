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

	_ "bctbackend/docs"
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

func (endpoint *listCashierSalesEndpoint) execute() {
	uriCashierId, ok := endpoint.extractCashierIdFromUri()
	if !ok {
		return
	}

	var saleSummaries []*models.SaleSummary
	if err := queries.GetCashierSales(endpoint.Database, uriCashierId, queries.CollectTo(&saleSummaries), queries.AllRows()); err != nil {
		endpoint.Logger.InternalError("Failed to retrieve cashier sales for user %d: %v", uriCashierId, err)
		failure_response.Unknown(endpoint.Context, "Could not retrieve cashier sales: "+err.Error())
		return
	}

	successResponse := ListCashierSalesSuccessResponse{
		Sales: algorithms.Map(saleSummaries, func(saleSummary *models.SaleSummary) *ListCashierSaleData {
			return endpoint.convertSaleSummaryToData(saleSummary)
		}),
	}

	endpoint.Context.IndentedJSON(http.StatusOK, successResponse)
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
	if err := endpoint.Context.ShouldBindUri(&uriParameters); err != nil {
		endpoint.Logger.InvalidInput("Failed to bind URI parameters: %v", err)
		failure_response.InvalidUriParameters(endpoint.Context, err.Error())
		return 0, false
	}

	uriUserId, err := models.ParseId(uriParameters.CashierId)
	if err != nil {
		endpoint.Logger.InvalidInput("Failed to parse cashier ID \"%s\" from URI: %v", uriParameters.CashierId, err)
		failure_response.InvalidUserId(endpoint.Context, err.Error())
		return 0, false
	}

	if !endpoint.ensureUserHasPermission(uriUserId) {
		return 0, false
	}

	return uriUserId, true
}

func (endpoint *listCashierSalesEndpoint) ensureUserHasPermission(queriedUser models.Id) bool {
	user, err := queries.GetUserWithId(endpoint.Database, endpoint.UserId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should not happen, as the userId is from the logged-in user
			endpoint.Logger.Bug("Logged in user does not exist")
			failure_response.Unknown(endpoint.Context, "Bug: logged in user does not exist")
			return false
		}
		failure_response.Unknown(endpoint.Context, "Could not retrieve logged in user: "+err.Error())
		return false
	}

	if user.RoleId.IsAdmin() {
		return true
	}

	if user.RoleId.IsCashier() {
		loggedInUser := endpoint.UserId

		if loggedInUser != queriedUser {
			endpoint.Logger.InvalidRequest("User tried to access sales of cashier %d, but is not the owning cashier", queriedUser)
			failure_response.Forbidden(endpoint.Context, "wrong_role", "Only accessible to owning cashiers or admins")
			return false
		}

		return true
	}

	endpoint.Logger.InvalidRequest("User tried to access sales of cashier %d, but is not a cashier or admin", queriedUser)
	failure_response.Forbidden(endpoint.Context, "wrong_role", "Only accessible to owning cashiers or admins")
	return false
}
