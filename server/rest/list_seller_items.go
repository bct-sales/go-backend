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

type GetSellerItemsItemData struct {
	ItemId       models.ID           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryId   models.ID           `binding:"required" json:"categoryId"`
	SellerId     models.ID           `binding:"required" json:"sellerId"`
	Donation     bool                `binding:"required" json:"donation"`
	Charity      bool                `binding:"required" json:"charity"`
	Frozen       bool                `binding:"required" json:"frozen"`
}

type GetSellerItemsSuccessResponse struct {
	Items []*GetSellerItemsItemData `json:"items"`
}

// @Summary Get seller's items
// @Description Get a seller's items
// @Param seller_id path int true "Seller ID"
// @Produce json
// @Success 200 {object} GetSellerItemsSuccessResponse "Items successfully fetched"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to owning sellers and admins"
// @Failure 404 {object} failure_response.FailureResponse "No such user"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch items"
// @Router /seller/{seller_id}/items [get]
func GetSellerItems(arguments *HandlerFunctionArguments) {
	endpoint := getSellerItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type getSellerItemsEndpoint struct {
	Endpoint
}

func (ep *getSellerItemsEndpoint) execute() {
	if !ep.ensureUserHasRightRole() {
		return
	}

	queriedSellerId, sellerIdOk := ep.extractSellerIdFromURI()
	if !sellerIdOk {
		return
	}

	if !ep.ensureQueriedUserIsSeller(queriedSellerId) {
		return
	}
	if !ep.ensureUserHasPermissions(queriedSellerId) {
		return
	}

	itemSelection := ep.extractItemSelectionFromQueryParameters()

	items, itemsOk := ep.fetchSellerItemsFromDatabase(queriedSellerId, itemSelection)
	if !itemsOk {
		return
	}

	ep.sendSuccessResponse(items)
}

func (ep *getSellerItemsEndpoint) ensureUserHasRightRole() bool {
	if !ep.RoleId.IsSeller() && !ep.RoleId.IsAdmin() {
		ep.Logger.InvalidRequest("User lacks permissions to access seller items")
		failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to sellers and admins")
		return false
	}

	return true
}

func (ep *getSellerItemsEndpoint) extractSellerIdFromURI() (models.ID, bool) {
	var uriParameters struct {
		SellerId string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		ep.Logger.InvalidInput("Failed to bind URI parameters for GetSellerItems", "error", err)
		failure_response.InvalidUriParameters(ep.Context, err.Error())
		return 0, false
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)
	if err != nil {
		ep.Logger.InvalidInput("Failed to parse seller ID from URI", "error", err, "sellerId", uriParameters.SellerId)
		failure_response.InvalidUserId(ep.Context, err.Error())
		return 0, false
	}

	return uriSellerId, true
}

func (ep *getSellerItemsEndpoint) ensureQueriedUserIsSeller(queriedSellerId models.ID) bool {
	if err := queries.EnsureUserExistsAndHasRole(ep.Database, queriedSellerId, models.NewSellerRoleId()); err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			ep.Logger.InvalidRequest("Seller does not exist", "error", err, "sellerId", queriedSellerId)
			failure_response.UnknownUser(ep.Context, err.Error())
			return false
		}

		if errors.Is(err, dberr.ErrWrongRole) {
			ep.Logger.InvalidRequest("Can only list items of sellers", "nonSellerId", queriedSellerId)
			failure_response.WrongUser(ep.Context, "Can only list items of sellers")
			return false
		}

		ep.Logger.InternalError("Could not check user role", "error", err)
		failure_response.Unknown(ep.Context, "Could not check user role: "+err.Error())
		return false
	}

	return true
}

func (ep *getSellerItemsEndpoint) ensureUserHasPermissions(queriedSellerId models.ID) bool {
	if ep.UserId != queriedSellerId && !ep.RoleId.IsAdmin() {
		ep.Logger.InvalidRequest("Logged in user does not match URI seller ID", "uriSellerId", queriedSellerId)
		failure_response.WrongSeller(ep.Context, "Logged in user does not match URI seller ID")
		return false
	}

	return true
}

func (ep *getSellerItemsEndpoint) extractItemSelectionFromQueryParameters() queries.ItemSelection {
	switch ep.Context.Query("items") {
	case "all":
		return queries.AllItems
	case "hidden":
		return queries.OnlyHiddenItems
	default:
		return queries.OnlyVisibleItems
	}
}

func (ep *getSellerItemsEndpoint) fetchSellerItemsFromDatabase(queriedSellerId models.ID, itemSelection queries.ItemSelection) ([]*models.Item, bool) {
	items, err := queries.GetSellerItems(ep.Database, queriedSellerId, itemSelection)

	if err != nil {
		ep.Logger.InternalError("Could not retrieve seller items", "error", err, "queriedSellerId", queriedSellerId)
		failure_response.Unknown(ep.Context, "Could not retrieve seller items: "+err.Error())
		return nil, false
	}

	return items, true
}

func (ep *getSellerItemsEndpoint) sendSuccessResponse(items []*models.Item) {
	successResponse := GetSellerItemsSuccessResponse{Items: algorithms.Map(items, func(item *models.Item) *GetSellerItemsItemData {
		return &GetSellerItemsItemData{
			ItemId:       item.ItemID,
			AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryId:   item.CategoryID,
			SellerId:     item.SellerID,
			Donation:     item.Donation,
			Charity:      item.Charity,
			Frozen:       item.Frozen,
		}
	})}

	ep.Context.IndentedJSON(http.StatusOK, successResponse)
}
