package failure_response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FailureResponse struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

func BadRequest(context *gin.Context, message string) {
	response := &FailureResponse{Type: "bad_request", Details: message}
	context.JSON(http.StatusBadRequest, response)
}

func InvalidUserId(context *gin.Context, message string) {
	response := &FailureResponse{Type: "invalid_user_id", Details: message}
	context.JSON(http.StatusUnauthorized, response)
}

func InvalidItemId(context *gin.Context, message string) {
	response := &FailureResponse{Type: "invalid_item_id", Details: message}
	context.JSON(http.StatusUnauthorized, response)
}

func UnknownUser(context *gin.Context, message string) {
	response := &FailureResponse{Type: "unknown_user", Details: message}
	context.JSON(http.StatusUnauthorized, response)
}

func WrongPassword(context *gin.Context, message string) {
	response := &FailureResponse{Type: "wrong_password", Details: message}
	context.JSON(http.StatusUnauthorized, response)
}

func Unknown(context *gin.Context, message string) {
	response := &FailureResponse{Type: "unknown", Details: message}
	context.JSON(http.StatusInternalServerError, response)
}

func Forbidden(context *gin.Context, message string) {
	response := &FailureResponse{Type: "forbidden", Details: message}
	context.JSON(http.StatusForbidden, response)
}

func CannotUpdateFrozenItem(context *gin.Context, message string) {
	response := &FailureResponse{Type: "item_frozen", Details: message}
	context.JSON(http.StatusForbidden, response)
}

func InvalidPrice(context *gin.Context, message string) {
	response := &FailureResponse{Type: "invalid_price", Details: message}
	context.JSON(http.StatusBadRequest, response)
}