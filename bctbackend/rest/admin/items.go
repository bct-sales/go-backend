package admin

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
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
}

type GetItemsSuccessResponse struct {
	Items []GetItemsItemData `json:"items"`
}

type GetItemsFailureResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// @Summary Get list of items.
// @Description Returns all items. Only accessible to users with the admin role.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} GetItemsSuccessResponse "Items successfully fetched"
// @Failure 500 {object} GetItemsFailureResponse "Failed to fetch items"
// @Router /admin/items [get]
func GetItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		failureResponse := GetItemsFailureResponse{Type: "forbidden", Message: "Only accessible to admins"}
		context.JSON(http.StatusForbidden, failureResponse)
		return
	}

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items)); err != nil {
		failureResponse := GetItemsFailureResponse{Message: "Failed to fetch items"}
		context.JSON(http.StatusInternalServerError, failureResponse)
		return
	}

	response := GetItemsSuccessResponse{Items: algorithms.Map(items, func(item *models.Item) GetItemsItemData {
		return GetItemsItemData{
			ItemId:       item.ItemId,
			AddedAt:      rest.FromUnix(item.AddedAt),
			Description:  item.Description,
			PriceInCents: item.PriceInCents,
			CategoryId:   item.CategoryId,
			SellerId:     item.SellerId,
			Donation:     item.Donation,
			Charity:      item.Charity,
		}
	})}

	context.IndentedJSON(http.StatusOK, response)
}
