package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ListSellerItemsItemData struct {
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

type ListSellerItemsSuccessResponse struct {
	Items []*ListSellerItemsItemData `json:"items"`
}

func ListSellerItems(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := listSellerItemsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type listSellerItemsEndpoint struct {
	Endpoint
}

func (ep *listSellerItemsEndpoint) execute(context *gin.Context) {
	if !ep.ensureUserHasRightRole(context) {
		return
	}

	queriedSellerID, sellerIDOk := ep.extractSellerIDFromURI(context)
	if !sellerIDOk {
		return
	}

	if !ep.ensureQueriedUserIsSeller(context, queriedSellerID) {
		return
	}
	if !ep.ensureUserHasPermissions(context, queriedSellerID) {
		return
	}

	itemSelection := ep.extractItemSelectionFromQueryParameters(context)

	items, itemsOk := ep.fetchSellerItemsFromDatabase(context, queriedSellerID, itemSelection)
	if !itemsOk {
		return
	}

	ep.sendSuccessResponse(context, items)
}

func (ep *listSellerItemsEndpoint) ensureUserHasRightRole(context *gin.Context) bool {
	logger := ep.Logger

	if !ep.RoleID.IsSeller() && !ep.RoleID.IsAdmin() {
		logger.InvalidRequest("User lacks permissions to access seller items")

		failure_response.Forbidden(context, "wrong_role", "Only accessible to sellers and admins")

		return false
	}

	return true
}

func (ep *listSellerItemsEndpoint) extractSellerIDFromURI(context *gin.Context) (models.ID, bool) {
	logger := ep.Logger
	var uriParameters struct {
		SellerID string `binding:"required" uri:"id"`
	}

	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to bind URI parameters for GetSellerItems")

		failure_response.InvalidUriParameters(context, err.Error())

		return 0, false
	}

	uriSellerID, err := models.ParseID(uriParameters.SellerID)
	if err != nil {
		logger.AddInformation("sellerID", uriParameters.SellerID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse seller ID from URI")

		failure_response.InvalidUserID(context, err.Error())

		return 0, false
	}

	logger.AddInformation("queriedSellerID", uriSellerID)

	return uriSellerID, true
}

func (ep *listSellerItemsEndpoint) ensureQueriedUserIsSeller(context *gin.Context, queriedSellerID models.ID) bool {
	logger := ep.Logger
	ctx := context.Request.Context()

	if err := queries.EnsureUserExistsAndHasRole(ctx, ep.Database, queriedSellerID, models.NewSellerRoleID()); err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			logger.InvalidRequest("Seller does not exist")

			failure_response.UnknownUser(context, err.Error())

			return false
		}

		if errors.Is(err, dberr.ErrWrongRole) {
			logger.InvalidRequest("Can only list items of sellers")

			failure_response.WrongUser(context, "Can only list items of sellers")

			return false
		}

		logger.InternalError("Could not check user role")

		failure_response.Unknown(context, "Could not check user role: "+err.Error())

		return false
	}

	return true
}

func (ep *listSellerItemsEndpoint) ensureUserHasPermissions(context *gin.Context, queriedSellerID models.ID) bool {
	logger := ep.Logger

	if ep.UserID != queriedSellerID && !ep.RoleID.IsAdmin() {
		logger.InvalidRequest("Logged in user does not match URI seller ID")

		failure_response.WrongSeller(context, "Logged in user does not match URI seller ID")

		return false
	}

	return true
}

func (ep *listSellerItemsEndpoint) extractItemSelectionFromQueryParameters(context *gin.Context) queries.ItemSelection {
	switch context.Query("items") {
	case "all":
		return queries.AllItems
	case "hidden":
		return queries.OnlyHiddenItems
	default:
		return queries.OnlyVisibleItems
	}
}

func (ep *listSellerItemsEndpoint) fetchSellerItemsFromDatabase(context *gin.Context, queriedSellerID models.ID, itemSelection queries.ItemSelection) ([]*models.Item, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	items, err := queries.GetSellerItems(ctx, ep.Database, queriedSellerID, itemSelection)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Could not retrieve seller items")

		failure_response.Unknown(context, "Could not retrieve seller items: "+err.Error())

		return nil, false
	}

	return items, true
}

func (ep *listSellerItemsEndpoint) sendSuccessResponse(context *gin.Context, items []*models.Item) {
	successResponse := ListSellerItemsSuccessResponse{Items: algorithms.Map(items, func(item *models.Item) *ListSellerItemsItemData {
		return &ListSellerItemsItemData{
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

	context.IndentedJSON(http.StatusOK, successResponse)
}
