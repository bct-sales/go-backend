package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"errors"
	"net/http"
)

type AddSellerItemPayload struct {
	Price       *models.MoneyInCents `binding:"required" json:"priceInCents"`
	Description *string              `binding:"required" json:"description"`
	CategoryId  models.Id            `binding:"required" json:"categoryId"`
	Donation    *bool                `binding:"required" json:"donation"` // needs to be a pointer to differentiate between false and not present
	Charity     *bool                `binding:"required" json:"charity"`  // needs to be a pointer to differentiate between false and not present
}

type AddSellerItemResponse struct {
	ItemId models.Id `json:"itemId"`
}

// @Summary Add an item as seller
// @Description Add an item as a seller
// @Param seller_id path int true "Seller ID"
// @Produce json
// @Success 200 {object} AddSellerItemResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to sellers and admins, or invalid item data"
// @Failure 404 {object} failure_response.FailureResponse "No such user or category"
// @Failure 500 {object} failure_response.FailureResponse "Failed to add item"
// @Router /seller/{seller_id}/items [put]
func AddSellerItem(arguments *HandlerFunctionArguments) {
	endpoint := addSellerItemEndpoint{
		HandlerFunctionArguments: *arguments,
	}

	endpoint.execute()
}

type addSellerItemEndpoint struct {
	HandlerFunctionArguments
}

func (ep *addSellerItemEndpoint) execute() {
	if !ep.ensureUserIsSeller() {
		return
	}

	var uriParameters struct {
		SellerId string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		ep.Logger.InvalidInput("Failed to parse URI parameters", "error", err, "uriParameters", uriParameters)
		failure_response.InvalidUriParameters(ep.Context, err.Error())
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)
	if err != nil {
		ep.Logger.InvalidInput("Failed to parse seller ID from URI", "error", err, "sellerId", uriParameters.SellerId)
		failure_response.InvalidUserId(ep.Context, err.Error())
		return
	}

	{
		sellerExists, err := queries.UserWithIdExists(ep.Database, uriSellerId)
		if err != nil {
			ep.Logger.InternalError("Failed to check if seller exists", "error", err, "sellerId", uriSellerId)
			failure_response.Unknown(ep.Context, err.Error())
			return
		}
		if !sellerExists {
			ep.Logger.InvalidRequest("Blocked attempt to add item for non-existing seller", "sellerId", uriSellerId)
			failure_response.UnknownUser(ep.Context, "Seller does not exist")
			return
		}
	}

	if uriSellerId != ep.UserId {
		ep.Logger.InvalidRequest("Blocked attempt to add item for different seller", "uriSellerId", uriSellerId)
		failure_response.WrongSeller(ep.Context, "Logged in user does not match URI seller ID")
		return
	}

	var payload AddSellerItemPayload
	if err := ep.Context.ShouldBindJSON(&payload); err != nil {
		ep.Logger.InvalidInput("Failed to parse AddSellerItem payload", "error", err, "payload", payload)
		failure_response.InvalidRequest(ep.Context, err.Error())
		return
	}

	timestamp := ep.Clock.Now()
	itemId, err := queries.AddItem(
		ep.Database,
		timestamp,
		*payload.Description,
		*payload.Price,
		payload.CategoryId,
		ep.UserId,
		*payload.Donation,
		*payload.Charity,
		false,
		false,
	)

	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchCategory) {
			ep.Logger.InvalidRequest("Blocked attempt to add item with unknown category", "categoryId", payload.CategoryId)
			failure_response.UnknownCategory(ep.Context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrNoSuchUser) {
			ep.Logger.InvalidRequest("Blocked attempt to add item for non-existing user", "userId", ep.UserId)
			failure_response.UnknownUser(ep.Context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrWrongRole) {
			ep.Logger.Bug("[BUG] Failed to add item to non-seller; this error should have been caught earlier")
			failure_response.Unknown(ep.Context, "Bug: this error should not happen")
			return
		}

		if errors.Is(err, dberr.ErrInvalidPrice) {
			ep.Logger.InvalidRequest("Blocked attempt to add item with invalid price", "price", payload.Price)
			failure_response.InvalidPrice(ep.Context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrInvalidItemDescription) {
			ep.Logger.InvalidRequest("Blocked attempt to add item with invalid description", "description", payload.Description)
			failure_response.InvalidItemDescription(ep.Context, err.Error())
			return
		}

		ep.Logger.InternalError("Failed to add seller item", "error", err)
		failure_response.Unknown(ep.Context, err.Error())
		return
	}

	response := AddSellerItemResponse{ItemId: itemId}
	ep.Context.JSON(http.StatusCreated, response)
}

func (ep *addSellerItemEndpoint) ensureUserIsSeller() bool {
	if !ep.RoleId.IsSeller() {
		ep.Logger.InvalidRequest("Blocked attempt to add item with wrong role")
		failure_response.WrongRole(ep.Context, "Must be seller to add item")
		return false
	}

	return true
}
