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

type GetSaleInformationSuccessResponse struct {
	SaleID          models.ID          `binding:"required" json:"saleId"`
	CashierID       models.ID          `binding:"required" json:"cashierId"`
	TransactionTime rest.DateTime      `binding:"required" json:"transactionTime"`
	Items           []*GetSaleItemData `binding:"required" json:"items"`
}

type GetSaleItemData struct {
	ItemID       models.ID           `binding:"required" json:"itemId"`
	SellerID     models.ID           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryID   models.ID           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	Large        *bool               `binding:"required" json:"large"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
}

type getSaleInformationEndpoint struct {
	Endpoint
}

func GetSaleInformation(arguments *HandlerFunctionArguments) {
	endpoint := &getSaleInformationEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

func (ep *getSaleInformationEndpoint) execute() {
	logger := ep.Logger

	if !ep.ensureUserHasRightRole() {
		return
	}

	saleID, ok := ep.extractSaleIDFromUri()
	if !ok {
		return
	}

	sale, err := queries.GetSaleWithID(ep.Database, saleID)
	if err != nil {
		logger.AddInformation("saleID", saleID)

		if errors.Is(err, dberr.ErrNoSuchSale) {
			logger.InvalidRequest("No such sale found")

			failure_response.UnknownSale(ep.Context, err.Error())

			return
		}

		logger.AddInformation("error", err)
		logger.InternalError("Could not retrieve sale information")

		failure_response.Unknown(ep.Context, "Could not retrieve sale information: "+err.Error())

		return
	}

	if ep.RoleID.IsCashier() && sale.CashierID != ep.UserID {
		logger.AddInformation("saleID", saleID)
		logger.AddInformation("sale owner ID", sale.CashierID)
		logger.InvalidRequest("Sale is not owned by the cashier")

		failure_response.Forbidden(ep.Context, "wrong_sale", "Only accessible to cashiers and owning cashiers")

		return
	}

	saleItems, err := queries.GetSaleItems(ep.Database, saleID)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSale) {
			logger.AddInformation("saleID", saleID)
			logger.Bug("No such sale found; should have been caught earlier")
			failure_response.UnknownSale(ep.Context, err.Error())
			return
		}

		logger.AddInformation("saleID", saleID)
		logger.AddInformation("error", err)
		logger.InternalError("Could not retrieve sale items")

		failure_response.Unknown(ep.Context, "Could not retrieve sale information: "+err.Error())

		return
	}

	response := GetSaleInformationSuccessResponse{
		SaleID:          sale.SaleID,
		CashierID:       sale.CashierID,
		TransactionTime: rest.ConvertTimestampToDateTime(sale.TransactionTime),
		Items:           algorithms.Map(saleItems, ep.convertSaleItemToData),
	}

	ep.Context.JSON(http.StatusOK, response)
}

func (ep *getSaleInformationEndpoint) convertSaleItemToData(saleItem *models.Item) *GetSaleItemData {
	return &GetSaleItemData{
		ItemID:       saleItem.ItemID,
		SellerID:     saleItem.SellerID,
		Description:  saleItem.Description,
		PriceInCents: saleItem.PriceInCents,
		CategoryID:   saleItem.CategoryID,
		Charity:      &saleItem.Charity,
		Donation:     &saleItem.Donation,
		Large:        &saleItem.Large,
		AddedAt:      rest.ConvertTimestampToDateTime(saleItem.AddedAt),
	}
}

func (ep *getSaleInformationEndpoint) ensureUserHasRightRole() bool {
	logger := ep.Logger

	if !ep.RoleID.IsAdmin() && !ep.RoleID.IsCashier() {
		logger.InvalidRequest("User does not have the right role to access sale information")
		failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to cashiers and owning cashiers")
		return false
	}

	return true
}

func (ep *getSaleInformationEndpoint) extractSaleIDFromUri() (models.ID, bool) {
	logger := ep.Logger

	var uriParameters struct {
		SaleID string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Invalid URI parameters")

		failure_response.InvalidUriParameters(ep.Context, "Invalid URI parameters: "+err.Error())

		return 0, false
	}

	saleID, err := models.ParseID(uriParameters.SaleID)
	if err != nil {
		logger.AddInformation("saleID", uriParameters.SaleID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Invalid sale ID")

		failure_response.InvalidSaleID(ep.Context, err.Error())

		return 0, false
	}

	return saleID, true
}
