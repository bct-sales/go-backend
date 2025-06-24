package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetItemsItemData struct {
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

type GetItemsSuccessResponse struct {
	Items          []GetItemsItemData `json:"items"`
	TotalItemCount int                `json:"totalItemCount"`
}

// @Summary List all items of all sellers.
// @Description Returns all items of all sellers. Only accessible to users with the admin role.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} GetItemsSuccessResponse "Items successfully fetched"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch items"
// @Router /items [get]
func GetAllItems(context *gin.Context, configuration *configuration.Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	if roleId != models.NewAdminRoleId() {
		failure_response.WrongRole(context, "Only admins can list all items")
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

	var limit *int
	limitString := context.Query("limit")
	if limitString != "" {
		parsedLimit, err := strconv.Atoi(limitString)

		if err != nil {
			slog.Error("Failed to parse limit", "error", err)
			failure_response.BadRequest(context, "invalid_uri_parameters", "Failed to parse limit: "+err.Error())
			return
		}

		limit = &parsedLimit
	} else {
		limit = nil
	}

	var offset *int = nil
	offsetString := context.Query("offset")
	if offsetString != "" {
		parsedOffset, err := strconv.Atoi(offsetString)

		if err != nil {
			slog.Error("Failed to parse offset", "error", err)
			failure_response.BadRequest(context, "invalid_uri_parameters", "Failed to parse offset: "+err.Error())
			return
		}

		offset = &parsedOffset
	} else {
		offset = nil
	}

	rowSelection := queries.RowSelection(offset, limit)

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items), itemSelection, rowSelection); err != nil {
		slog.Error("Failed to get items", "error", err)
		failure_response.Unknown(context, "Failed to get items: "+err.Error())
		return
	}

	switch context.Query("format") {
	case "":
		items := algorithms.Map(items, func(item *models.Item) GetItemsItemData {
			return GetItemsItemData{
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
		})
		itemCount, err := queries.CountItems(db, itemSelection)
		if err != nil {
			slog.Error("Failed to count items", "error", err)
			failure_response.Unknown(context, "Failed to count items: "+err.Error())
			return
		}

		response := GetItemsSuccessResponse{
			Items:          items,
			TotalItemCount: itemCount,
		}

		context.IndentedJSON(http.StatusOK, response)
		return

	case "json":
		context.Header("Content-Type", "application/json")
		context.Header("Content-Disposition", "attachment; filename=\"items.json\"")
		context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		context.Header("Pragma", "no-cache")

		context.IndentedJSON(http.StatusOK, items)
		return

	case "csv":
		categoryNameTable, err := queries.GetCategoryNameTable(db)
		if err != nil {
			failure_response.Unknown(context, "Failed to get category map: "+err.Error())
			return
		}

		context.Header("Content-Type", "text/csv")
		context.Header("Content-Disposition", "attachment; filename=\"items.csv\"")
		context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		context.Header("Pragma", "no-cache")

		buffer := new(bytes.Buffer)
		if err := csv.FormatItemsAsCSV(items, categoryNameTable, buffer); err != nil {
			failure_response.Unknown(context, "Failed to format items as CSV: "+err.Error())
			return
		}
		string := buffer.String()
		context.String(http.StatusOK, string)
		return

	default:
		failure_response.Unknown(context, "Unknown format: "+context.Query("format"))
		return
	}
}
