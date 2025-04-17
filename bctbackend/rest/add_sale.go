package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddSalePayload struct {
	Items []models.Id `json:"item_ids" binding:"required"`
}

type AddSaleSuccessResponse struct {
	SaleId models.Id `json:"sale_id"`
}

// @Summary Add a new sale
// @Description Adds a new sale to the database. Only accessible to users with the cashier role.
// @Tags sales
// @Accept json
// @Produce json
// @Param AddSalePayload body AddSalePayload true "Payload containing item IDs"
// @Success 201 {object} AddSaleSuccessResponse "Sale successfully added"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to cashiers"
// @Failure 404 {object} failure_response.FailureResponse "Unknown item in sale"
// @Router /sales [post]
func AddSale(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		failure_response.WrongRole(context, "Adding sale is only accessible to cashiers")
		return
	}

	var payload AddSalePayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())
		return
	}

	timestamp := models.Now()

	saleId, err := queries.AddSale(
		db,
		userId,
		timestamp,
		payload.Items,
	)
	if err != nil {
		{
			var saleMissingItemsError *queries.SaleMissingItemsError
			if errors.As(err, &saleMissingItemsError) {
				failure_response.SaleMissingItems(context, err.Error())
				return
			}
		}

		{
			var duplicateItemInSaleError *queries.DuplicateItemInSaleError
			if errors.As(err, &duplicateItemInSaleError) {
				failure_response.DuplicateItemInSale(context, err.Error())
				return
			}
		}

		{
			var NoSuchItemError *queries.NoSuchItemError
			if errors.As(err, &NoSuchItemError) {
				failure_response.UnknownItem(context, err.Error())
				return
			}
		}

		failure_response.Unknown(context, "Failed to add sale: "+err.Error())
		return
	}

	response := AddSaleSuccessResponse{SaleId: saleId}
	context.JSON(http.StatusCreated, response)
}
