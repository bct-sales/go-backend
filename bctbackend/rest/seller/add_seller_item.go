package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddSellerItemPayload struct {
	Price       models.MoneyInCents `json:"price_in_cents" binding:"required"`
	Description string              `json:"description" binding:"required"`
	CategoryId  models.Id           `json:"category_id" binding:"required"`
	Donation    *bool               `json:"donation" binding:"required"` // needs to be a pointer to differentiate between false and not present
	Charity     *bool               `json:"charity" binding:"required"`  // needs to be a pointer to differentiate between false and not present
}

type AddSellerItemResponse struct {
	ItemId models.Id `json:"item_id"`
}

func AddSellerItem(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.SellerRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to sellers"})
		return
	}

	var uriParameters struct {
		SellerId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid URI parameters: " + err.Error()})
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse seller Id: " + err.Error()})
		return
	}

	if uriSellerId != userId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Logged in user does not match URI seller ID"})
		return
	}

	var payload AddSellerItemPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse payload: " + err.Error()})
		return
	}

	timestamp := models.Now()

	itemId, err := queries.AddItem(
		db,
		timestamp,
		payload.Description,
		payload.Price,
		payload.CategoryId,
		userId,
		*payload.Donation,
		*payload.Charity,
	)

	// TODO recognize error (e.g., if the category does not exist) and return StatusBadRequest or InternalServerError depending on the error
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Failed to add item"})
		return
	}

	response := AddSellerItemResponse{ItemId: itemId}
	context.JSON(http.StatusCreated, response)
}
