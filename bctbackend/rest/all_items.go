package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetItemsItemData struct {
	ItemId       models.Id                `json:"itemId"`
	AddedAt      rest.StructuredTimestamp `json:"addedAt"`
	Description  string                   `json:"description"`
	PriceInCents models.MoneyInCents      `json:"priceInCents"`
	CategoryId   models.Id                `json:"categoryId"`
	SellerId     models.Id                `json:"sellerId"`
	Donation     bool                     `json:"donation"`
	Charity      bool                     `json:"charity"`
	Frozen       bool                     `json:"frozen"`
}

type GetItemsSuccessResponse struct {
	Items []GetItemsItemData `json:"items"`
}

// @Summary List all items of all sellers.
// @Description Returns all items of all sellers. Only accessible to users with the admin role.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} GetItemsSuccessResponse "Items successfully fetched"
// @Failure 500 {object} GetItemsFailureResponse "Failed to fetch items"
// @Router /items [get]
func GetAllItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		failure_response.WrongRole(context, "Only admins can list all items")
		return
	}

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items)); err != nil {
		failure_response.Unknown(context, "Failed to get items: "+err.Error())
		return
	}

	response := GetItemsSuccessResponse{Items: algorithms.Map(items, func(item *models.Item) GetItemsItemData {
		return GetItemsItemData{
			ItemId:       item.ItemId,
			AddedAt:      rest.FromTimestamp(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryId:   item.CategoryId,
			SellerId:     item.SellerId,
			Donation:     item.Donation,
			Charity:      item.Charity,
			Frozen:       item.Frozen,
		}
	})}

	context.IndentedJSON(http.StatusOK, response)
}
