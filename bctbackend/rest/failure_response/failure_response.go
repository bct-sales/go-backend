package failure_response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FailureResponse struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

func BadRequest(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusBadRequest, response)
}

// User was not authenticated
func Unauthorized(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusUnauthorized, response)
}

// User was authenticated, but is not authorized to perform the action
func Forbidden(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusForbidden, response)
}

func NotFound(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusNotFound, response)
}

func Unknown(context *gin.Context, message string) {
	response := &FailureResponse{Type: "unknown", Details: message}
	context.JSON(http.StatusInternalServerError, response)
}

// Could not parse request
func InvalidRequest(context *gin.Context, message string) {
	BadRequest(context, "invalid_request", "invalid request: "+message)
}

func InvalidUriParameters(context *gin.Context, message string) {
	BadRequest(context, "invalid_uri_parameters", "invalid URI parameters: "+message)
}

// Ill-formed user ID, e.g., "abc" instead of "123"
func InvalidUserId(context *gin.Context, message string) {
	BadRequest(context, "invalid_user_id", "invalid user id: "+message)
}

// Ill-formed item ID, e.g., "abc" instead of "123"
func InvalidItemId(context *gin.Context, message string) {
	BadRequest(context, "invalid_item_id", "invalid item id: "+message)
}

// There is no item with the given ID
func UnknownItem(context *gin.Context, message string) {
	NotFound(context, "no_such_item", message)
}

// There is no user with the given ID
func UnknownUser(context *gin.Context, message string) {
	NotFound(context, "no_such_user", message)
}

func WrongUser(context *gin.Context, message string) {
	Forbidden(context, "wrong_user", message)
}

func UnknownCategory(context *gin.Context, message string) {
	NotFound(context, "no_such_category", message)
}

func WrongPassword(context *gin.Context, message string) {
	Unauthorized(context, "wrong_password", message)
}

func CannotUpdateFrozenItem(context *gin.Context, message string) {
	Forbidden(context, "item_frozen", message)
}

func InvalidPrice(context *gin.Context, message string) {
	BadRequest(context, "invalid_price", message)
}

func WrongRole(context *gin.Context, message string) {
	Forbidden(context, "wrong_role", message)
}

func DuplicateItemInSale(context *gin.Context, message string) {
	Forbidden(context, "duplicate_item_in_sale", message)
}

// No items in sale
func SaleMissingItems(context *gin.Context, message string) {
	Forbidden(context, "sale_missing_items", message)
}

// Seller trying to access other seller's data
func WrongSeller(context *gin.Context, message string) {
	Forbidden(context, "wrong_seller", message)
}

func InvalidItemDescription(context *gin.Context, message string) {
	BadRequest(context, "invalid_description", message)
}

func MissingSessionId(context *gin.Context, message string) {
	Unauthorized(context, "missing_session_id", message)
}

func UnknownSession(context *gin.Context, message string) {
	Unauthorized(context, "no_such_session", message)
}
