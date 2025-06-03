package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"errors"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetSellerItemsItemData struct {
	ItemId       models.Id           `json:"itemId"`
	AddedAt      rest.DateTime       `json:"addedAt"`
	Description  string              `json:"description"`
	PriceInCents models.MoneyInCents `json:"priceInCents"`
	CategoryId   models.Id           `json:"categoryId"`
	SellerId     models.Id           `json:"sellerId"`
	Donation     bool                `json:"donation"`
	Charity      bool                `json:"charity"`
	Frozen       bool                `json:"frozen"`
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
func GetSellerItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.SellerRoleId && roleId != models.AdminRoleId {
		failure_response.Forbidden(context, "wrong_role", "Only accessible to sellers and admins")
		return
	}

	var uriParameters struct {
		SellerId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(context, err.Error())
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)
	if err != nil {
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	if err := queries.EnsureUserExistsAndHasRole(db, uriSellerId, models.SellerRoleId); err != nil {
		if errors.Is(err, database.ErrNoSuchUser) {
			failure_response.UnknownUser(context, err.Error())
			return
		}

		if errors.Is(err, database.ErrWrongRole) {
			failure_response.WrongUser(context, "Can only list items of sellers")
			return
		}

		failure_response.Unknown(context, "Could not check user role: "+err.Error())
		return
	}

	if userId != uriSellerId && roleId != models.AdminRoleId {
		failure_response.WrongSeller(context, "Logged in user does not match URI seller ID")
		return
	}

	var itemSelection queries.ItemSelection
	switch context.Query("items") {
	case "all":
		itemSelection = queries.AllItems
	case "hidden":
		itemSelection = queries.OnlyHiddenItems
	default:
		itemSelection = queries.OnlyVisibleItems
	}

	items, err := queries.GetSellerItems(db, uriSellerId, itemSelection)
	if err != nil {
		failure_response.Unknown(context, "Could not retrieve seller items: "+err.Error())
		return
	}

	successResponse := GetSellerItemsSuccessResponse{Items: algorithms.Map(items, func(item *models.Item) *GetSellerItemsItemData {
		return &GetSellerItemsItemData{
			ItemId:       item.ItemID,
			AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryId:   item.CategoryID,
			SellerId:     item.SellerId,
			Donation:     item.Donation,
			Charity:      item.Charity,
			Frozen:       item.Frozen,
		}
	})}

	context.IndentedJSON(http.StatusOK, successResponse)
}
