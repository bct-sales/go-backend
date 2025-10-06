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
	SaleId          models.ID          `binding:"required" json:"saleId"`
	CashierId       models.ID          `binding:"required" json:"cashierId"`
	TransactionTime rest.DateTime      `binding:"required" json:"transactionTime"`
	Items           []*GetSaleItemData `binding:"required" json:"items"`
}

type GetSaleItemData struct {
	ItemId       models.ID           `binding:"required" json:"itemId"`
	SellerId     models.ID           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryId   models.ID           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
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

func (endpoint *getSaleInformationEndpoint) execute() {
	logger := endpoint.Logger

	if !endpoint.ensureUserHasRightRole() {
		return
	}

	saleId, ok := endpoint.extractSaleIdFromUri()
	if !ok {
		return
	}

	sale, err := queries.GetSaleWithId(endpoint.Database, saleId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSale) {
			logger.InvalidRequest("No such sale found", "saleId", saleId)
			failure_response.UnknownSale(endpoint.Context, err.Error())
			return
		}

		logger.InternalError("Could not retrieve sale information", "saleId", saleId, "error", err)
		failure_response.Unknown(endpoint.Context, "Could not retrieve sale information: "+err.Error())
		return
	}

	if endpoint.RoleId.IsCashier() && sale.CashierID != endpoint.UserId {
		logger.InvalidRequest("Sale is not owned by the cashier", "saleId", saleId, "saleOwnerId", sale.CashierID)
		failure_response.Forbidden(endpoint.Context, "wrong_sale", "Only accessible to cashiers and owning cashiers")
		return
	}

	saleItems, err := queries.GetSaleItems(endpoint.Database, saleId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSale) {
			logger.Bug("No such sale found; should have been caught earlier", "saleId", saleId)
			failure_response.UnknownSale(endpoint.Context, err.Error())
			return
		}

		logger.InternalError("Could not retrieve sale items", "saleId", saleId, "error", err)
		failure_response.Unknown(endpoint.Context, "Could not retrieve sale information: "+err.Error())
		return
	}

	response := GetSaleInformationSuccessResponse{
		SaleId:          sale.SaleID,
		CashierId:       sale.CashierID,
		TransactionTime: rest.ConvertTimestampToDateTime(sale.TransactionTime),
		Items:           algorithms.Map(saleItems, endpoint.convertSaleItemToData),
	}

	endpoint.Context.JSON(http.StatusOK, response)
}

func (endpoint *getSaleInformationEndpoint) convertSaleItemToData(saleItem *models.Item) *GetSaleItemData {
	return &GetSaleItemData{
		ItemId:       saleItem.ItemID,
		SellerId:     saleItem.SellerID,
		Description:  saleItem.Description,
		PriceInCents: saleItem.PriceInCents,
		CategoryId:   saleItem.CategoryID,
		Charity:      &saleItem.Charity,
		Donation:     &saleItem.Donation,
		AddedAt:      rest.ConvertTimestampToDateTime(saleItem.AddedAt),
	}
}

func (endpoint *getSaleInformationEndpoint) ensureUserHasRightRole() bool {
	if !endpoint.RoleId.IsAdmin() && !endpoint.RoleId.IsCashier() {
		endpoint.Logger.InvalidRequest("User does not have the right role to access sale information")
		failure_response.Forbidden(endpoint.Context, "wrong_role", "Only accessible to cashiers and owning cashiers")
		return false
	}

	return true
}

func (endpoint *getSaleInformationEndpoint) extractSaleIdFromUri() (models.ID, bool) {
	var uriParameters struct {
		SaleId string `binding:"required" uri:"id"`
	}
	if err := endpoint.Context.ShouldBindUri(&uriParameters); err != nil {
		endpoint.Logger.InvalidInput("Invalid URI parameters", "error", err)
		failure_response.InvalidUriParameters(endpoint.Context, "Invalid URI parameters: "+err.Error())
		return 0, false
	}

	saleId, err := models.ParseId(uriParameters.SaleId)
	if err != nil {
		endpoint.Logger.InvalidInput("Invalid sale ID", "saleId", uriParameters.SaleId, "error", err)
		failure_response.InvalidSaleId(endpoint.Context, err.Error())
		return 0, false
	}

	return saleId, true
}
