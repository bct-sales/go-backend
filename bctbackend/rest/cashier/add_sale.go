package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
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

type AddSaleFailureResponse struct {
	Message string `json:"message"`
}

// @Summary Add a new sale
// @Description Adds a new sale to the database. Only accessible to users with the cashier role.
// @Tags sales
// @Accept json
// @Produce json
// @Param AddSalePayload body AddSalePayload true "Payload containing item IDs"
// @Success 201 {object} AddSaleSuccessResponse "Sale successfully added"
// @Failure 400 {object} AddSaleFailureResponse "Failed to parse payload or add sale"
// @Failure 403 {object} AddSaleFailureResponse "Only accessible to cashiers"
// @Router /sales [post]
func AddSale(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		errorResponse := AddSaleFailureResponse{Message: "Only accessible to cashiers"}
		context.JSON(http.StatusForbidden, errorResponse)
		return
	}

	var payload AddSalePayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		errorResponse := AddSaleFailureResponse{Message: "Failed to parse payload: " + err.Error()}
		context.JSON(http.StatusBadRequest, errorResponse)
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
		var duplicateItemInSaleError *queries.DuplicateItemInSaleError
		if errors.As(err, &duplicateItemInSaleError) {
			errorResponse := AddSaleFailureResponse{Message: "Duplicate item in sale"}
			context.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		errorResponse := AddSaleFailureResponse{Message: "Failed to add sale"}
		context.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	response := AddSaleSuccessResponse{SaleId: saleId}
	context.JSON(http.StatusCreated, response)
}
