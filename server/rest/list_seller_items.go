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
	ItemID       models.ID           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryID   models.ID           `binding:"required" json:"categoryId"`
	SellerID     models.ID           `binding:"required" json:"sellerId"`
	Donation     bool                `binding:"required" json:"donation"`
	Charity      bool                `binding:"required" json:"charity"`
	Large        bool                `binding:"required" json:"large"`
	Frozen       bool                `binding:"required" json:"frozen"`
}

type GetSellerItemsSuccessResponse struct {
	Items []*GetSellerItemsItemData `json:"items"`
}

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

	queriedSellerID, sellerIDOk := ep.extractSellerIDFromURI()
	if !sellerIDOk {
		return
	}

	if !ep.ensureQueriedUserIsSeller(queriedSellerID) {
		return
	}
	if !ep.ensureUserHasPermissions(queriedSellerID) {
		return
	}

	itemSelection := ep.extractItemSelectionFromQueryParameters()

	items, itemsOk := ep.fetchSellerItemsFromDatabase(queriedSellerID, itemSelection)
	if !itemsOk {
		return
	}

	ep.sendSuccessResponse(items)
}

func (ep *getSellerItemsEndpoint) ensureUserHasRightRole() bool {
	logger := ep.Logger

	if !ep.RoleID.IsSeller() && !ep.RoleID.IsAdmin() {
		logger.InvalidRequest("User lacks permissions to access seller items")

		failure_response.Forbidden(ep.Context, "wrong_role", "Only accessible to sellers and admins")

		return false
	}

	return true
}

func (ep *getSellerItemsEndpoint) extractSellerIDFromURI() (models.ID, bool) {
	logger := ep.Logger
	var uriParameters struct {
		SellerID string `binding:"required" uri:"id"`
	}

	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to bind URI parameters for GetSellerItems")

		failure_response.InvalidUriParameters(ep.Context, err.Error())

		return 0, false
	}

	uriSellerID, err := models.ParseID(uriParameters.SellerID)
	if err != nil {
		logger.AddInformation("sellerID", uriParameters.SellerID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse seller ID from URI")

		failure_response.InvalidUserID(ep.Context, err.Error())

		return 0, false
	}

	logger.AddInformation("queriedSellerID", uriSellerID)

	return uriSellerID, true
}

func (ep *getSellerItemsEndpoint) ensureQueriedUserIsSeller(queriedSellerID models.ID) bool {
	logger := ep.Logger

	if err := queries.EnsureUserExistsAndHasRole(ep.Database, queriedSellerID, models.NewSellerRoleID()); err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			logger.InvalidRequest("Seller does not exist")

			failure_response.UnknownUser(ep.Context, err.Error())

			return false
		}

		if errors.Is(err, dberr.ErrWrongRole) {
			logger.InvalidRequest("Can only list items of sellers")

			failure_response.WrongUser(ep.Context, "Can only list items of sellers")

			return false
		}

		logger.InternalError("Could not check user role")

		failure_response.Unknown(ep.Context, "Could not check user role: "+err.Error())

		return false
	}

	return true
}

func (ep *getSellerItemsEndpoint) ensureUserHasPermissions(queriedSellerID models.ID) bool {
	logger := ep.Logger

	if ep.UserID != queriedSellerID && !ep.RoleID.IsAdmin() {
		logger.InvalidRequest("Logged in user does not match URI seller ID")

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

func (ep *getSellerItemsEndpoint) fetchSellerItemsFromDatabase(queriedSellerID models.ID, itemSelection queries.ItemSelection) ([]*models.Item, bool) {
	logger := ep.Logger
	items, err := queries.GetSellerItems(ep.Database, queriedSellerID, itemSelection)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Could not retrieve seller items")

		failure_response.Unknown(ep.Context, "Could not retrieve seller items: "+err.Error())

		return nil, false
	}

	return items, true
}

func (ep *getSellerItemsEndpoint) sendSuccessResponse(items []*models.Item) {
	successResponse := GetSellerItemsSuccessResponse{Items: algorithms.Map(items, func(item *models.Item) *GetSellerItemsItemData {
		return &GetSellerItemsItemData{
			ItemID:       item.ItemID,
			AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryID:   item.CategoryID,
			SellerID:     item.SellerID,
			Donation:     item.Donation,
			Charity:      item.Charity,
			Large:        item.Large,
			Frozen:       item.Frozen,
		}
	})}

	ep.Context.IndentedJSON(http.StatusOK, successResponse)
}
