package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetUserInformationItem struct {
	ItemId       models.Id           `json:"item_id" binding:"required"`
	AddedAt      models.Timestamp    `json:"added_at" binding:"required"`
	SellerId     models.Id           `json:"seller_id" binding:"required"`
	Description  string              `json:"description" binding:"required"`
	PriceInCents models.MoneyInCents `json:"price_in_cents" binding:"required"`
	CategoryId   models.Id           `json:"category_id" binding:"required"`
	Charity      *bool               `json:"charity" binding:"required"`
	Donation     *bool               `json:"donation" binding:"required"`
	Frozen       *bool               `json:"frozen" binding:"required"`
	SaleCount    *int                `json:"sale_count" binding:"required"`
}

type GetUserInformationSale struct {
	SaleId          models.Id        `json:"sale_id" binding:"required"`
	TransactionTime models.Timestamp `json:"transaction_time" binding:"required"`
}

type GetUserInformationSuccessResponse struct {
	Role         string            `json:"role" binding:"required"`
	Password     string            `json:"password" binding:"required"`
	CreatedAt    models.Timestamp  `json:"created_at" binding:"required"`
	LastActivity *models.Timestamp `json:"last_activity" binding:"required"`
}

type GetSellerInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
	Items *[]*GetUserInformationItem `json:"items" binding:"required"`
}

type GetAdminInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
}

type GetCashierInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
	Sales *[]*GetUserInformationSale `json:"sales" binding:"required"`
}

func convertItemToGetUserInformationItem(item *queries.ItemWithSaleCount) *GetUserInformationItem {
	return &GetUserInformationItem{
		ItemId:       item.ItemId,
		Description:  item.Description,
		SellerId:     item.SellerId,
		AddedAt:      item.AddedAt,
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryId,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Frozen:       &item.Frozen,
		SaleCount:    &item.SaleCount,
	}
}

func convertSaleToGetUserInformationSale(sale *models.Sale) *GetUserInformationSale {
	return &GetUserInformationSale{
		SaleId:          sale.SaleId,
		TransactionTime: sale.TransactionTime,
	}
}

// @Summary Get information about a user
// @Description Get information about a user.
// @Success 200 {object} GetItemInformationSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to admins"
// @Failure 404 {object} failure_response.FailureResponse "User not found"
// @Failure 500 {object} failure_response.FailureResponse "Internal server error"
// @Router /users/{id} [get]
func GetUserInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(context, "Invalid URI parameters: "+err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	user, err := queries.GetUserWithId(db, itemId)
	if err != nil {
		{
			var noSuchUserError *queries.NoSuchUserError
			if errors.As(err, &noSuchUserError) {
				failure_response.UnknownUser(context, err.Error())
				return
			}
		}

		failure_response.Unknown(context, err.Error())
		return
	}

	roleName, err := models.NameOfRole(user.RoleId)
	if err != nil {
		failure_response.Unknown(context, err.Error())
		return
	}

	if user.RoleId == models.AdminRoleId {
		response := GetAdminInformationSuccessResponse{
			GetUserInformationSuccessResponse: GetUserInformationSuccessResponse{
				Role:         roleName,
				Password:     user.Password,
				CreatedAt:    user.CreatedAt,
				LastActivity: user.LastActivity,
			},
		}
		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleId == models.SellerRoleId {
		items, err := queries.GetSellerItemsWithSaleCounts(db, user.UserId)
		if err != nil {
			{
				var noSuchUserError *queries.NoSuchUserError
				if errors.As(err, &noSuchUserError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			{
				var invalidRoleError *queries.InvalidRoleError
				if errors.As(err, &invalidRoleError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			failure_response.Unknown(context, fmt.Errorf("failed to find information about seller: %w", err).Error())
			return
		}

		convertedItems := algorithms.Map(items, convertItemToGetUserInformationItem)

		response := GetSellerInformationSuccessResponse{
			GetUserInformationSuccessResponse: GetUserInformationSuccessResponse{
				Role:         roleName,
				Password:     user.Password,
				CreatedAt:    user.CreatedAt,
				LastActivity: user.LastActivity,
			},
			Items: &convertedItems,
		}

		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleId == models.CashierRoleId {
		sales, err := queries.GetSalesWithCashier(db, user.UserId)
		if err != nil {
			{
				var noSuchUserError *queries.NoSuchUserError
				if errors.As(err, &noSuchUserError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			{
				var invalidRoleError *queries.InvalidRoleError
				if errors.As(err, &invalidRoleError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			failure_response.Unknown(context, err.Error())
			return
		}

		convertedSales := algorithms.Map(sales, convertSaleToGetUserInformationSale)

		response := GetCashierInformationSuccessResponse{
			GetUserInformationSuccessResponse: GetUserInformationSuccessResponse{
				Role:         roleName,
				Password:     user.Password,
				CreatedAt:    user.CreatedAt,
				LastActivity: user.LastActivity,
			},
			Sales: &convertedSales,
		}
		context.JSON(http.StatusOK, response)
		return
	} else {
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %s", roleName))
		return
	}
}
