package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
	"net/http"
)

type GetItemInformationSuccessResponse struct {
	ItemId       models.Id           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	SellerId     models.Id           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryId   models.Id           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	Frozen       *bool               `binding:"required" json:"frozen"`
	SoldIn       *[]models.Id        `binding:"required" json:"soldIn"`
}

// @Summary Get information about an item
// @Description Get information about an item.
// @Success 200 {object} GetItemInformationSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to cashiers, admins and owner sellers"
// @Failure 404 {object} failure_response.FailureResponse "Item not found"
// @Router /items/{id} [get]
func GetItemInformation(arguments *HandlerFunctionArguments) {
	context := arguments.Context
	userId := arguments.UserId
	roleId := arguments.RoleId
	db := arguments.Database
	logger := arguments.Logger

	var uriParameters struct {
		ItemId string `binding:"required" uri:"id"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.InvalidInput("Failed to parse URI parameters", "error", err)
		failure_response.InvalidUriParameters(context, "Invalid URI parameters: "+err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		logger.InvalidInput("Failed to parse item ID", "error", err, "itemId", uriParameters.ItemId)
		failure_response.InvalidItemId(context, err.Error())
		return
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("Attempt to access a non-existing item", "itemId", itemId)
			failure_response.UnknownItem(context, err.Error())
			return
		}

		failure_response.Unknown(context, err.Error())
		return
	}

	if item.SellerID != userId && roleId.IsSeller() {
		logger.InvalidRequest("Blocked attempt to access item not owned by the seller", "itemId", item.ItemID, "itemUserId", item.SellerID)
		failure_response.WrongSeller(context, "Only the owning seller can access this item")
		return
	}

	soldIn, err := queries.GetSalesWithItem(db, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.Bug("Unknown item; should have been caught earlier", "itemId", itemId)
			failure_response.Unknown(context, "Bug: this should be caught by the previous query")
			return
		}

		logger.InternalError("Failed to get sales for item", "error", err, "itemId", itemId)
		failure_response.Unknown(context, err.Error())
		return
	}

	response := GetItemInformationSuccessResponse{
		ItemId:       item.ItemID,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		SellerId:     item.SellerID,
		Description:  item.Description,
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryID,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Frozen:       &item.Frozen,
		SoldIn:       &soldIn,
	}

	context.JSON(http.StatusOK, response)
}
