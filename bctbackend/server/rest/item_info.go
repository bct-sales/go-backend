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

type getItemInformationEndpoint struct {
	HandlerFunctionArguments
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
	endpoint := getItemInformationEndpoint{
		HandlerFunctionArguments: *arguments,
	}
	endpoint.execute()
}

func (ep *getItemInformationEndpoint) execute() {
	itemId, foundItemId := ep.retrieveItemIdFromUri()
	if !foundItemId {
		return
	}

	item := ep.retrieveItemFromDatabase(itemId)
	if item == nil {
		return
	}

	if item.SellerID != ep.UserId && ep.RoleId.IsSeller() {
		ep.Logger.InvalidRequest("Blocked attempt to access item not owned by the seller", "itemId", item.ItemID, "itemUserId", item.SellerID)
		failure_response.WrongSeller(ep.Context, "Only the owning seller can access this item")
		return
	}

	soldIn, err := queries.GetSalesWithItem(ep.Database, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			ep.Logger.Bug("Unknown item; should have been caught earlier", "itemId", itemId)
			failure_response.Unknown(ep.Context, "Bug: this should be caught by the previous query")
			return
		}

		ep.Logger.InternalError("Failed to get sales for item", "error", err, "itemId", itemId)
		failure_response.Unknown(ep.Context, err.Error())
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

	ep.Context.JSON(http.StatusOK, response)
}

func (ep *getItemInformationEndpoint) retrieveItemIdFromUri() (models.Id, bool) {
	var uriParameters struct {
		ItemId string `binding:"required" uri:"id"`
	}

	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		ep.Logger.InvalidInput("Failed to parse URI parameters", "error", err)
		failure_response.InvalidUriParameters(ep.Context, "Invalid URI parameters: "+err.Error())
		return 0, false
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		ep.Logger.InvalidInput("Failed to parse item ID", "error", err, "itemId", uriParameters.ItemId)
		failure_response.InvalidItemId(ep.Context, err.Error())
		return 0, false
	}

	return itemId, true
}

func (ep *getItemInformationEndpoint) retrieveItemFromDatabase(itemId models.Id) *models.Item {
	item, err := queries.GetItemWithId(ep.Database, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			ep.Logger.InvalidRequest("Attempt to access a non-existing item", "itemId", itemId)
			failure_response.UnknownItem(ep.Context, err.Error())
			return nil
		}

		failure_response.Unknown(ep.Context, err.Error())
		return nil
	}

	return item
}
