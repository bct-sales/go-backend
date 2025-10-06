package failure_response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FailureResponse struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

// HTTP status code 400
func BadRequest(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusBadRequest, response)
}

// User was not authenticated
// HTTP status code 401
func Unauthorized(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusUnauthorized, response)
}

// User was authenticated, but is not authorized to perform the action
// HTTP status code 403
func Forbidden(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusForbidden, response)
}

// HTTP status code 404
func NotFound(context *gin.Context, errorType string, message string) {
	response := &FailureResponse{Type: errorType, Details: message}
	context.JSON(http.StatusNotFound, response)
}

// HTTP status code 500
func Unknown(context *gin.Context, message string) {
	response := &FailureResponse{Type: "unknown", Details: message}
	context.JSON(http.StatusInternalServerError, response)
}

// Could not parse request
// HTTP status code 400
func InvalidRequest(context *gin.Context, message string) {
	BadRequest(context, "invalid_request", "invalid request: "+message)
}

// HTTP status code 400
func InvalidUriParameters(context *gin.Context, message string) {
	BadRequest(context, "invalid_uri_parameters", "invalid URI parameters: "+message)
}

// Ill-formed user ID, e.g., "abc" instead of "123"
// HTTP status code 400
func InvalidUserID(context *gin.Context, message string) {
	BadRequest(context, "invalid_user_id", "invalid user id: "+message)
}

// Ill-formed item ID, e.g., "abc" instead of "123".
// HTTP status code 400
func InvalidItemID(context *gin.Context, message string) {
	BadRequest(context, "invalid_item_id", "invalid item id: "+message)
}

// Ill-formed sale ID, e.g., "abc" instead of "123"
// HTTP status code 400
func InvalidSaleID(context *gin.Context, message string) {
	BadRequest(context, "invalid_sale_id", "invalid sale id: "+message)
}

// There is no item with the given ID
// HTTP status code 404
func UnknownItem(context *gin.Context, message string) {
	NotFound(context, "no_such_item", message)
}

// There is no user with the given ID
// HTTP status code 404
func UnknownUser(context *gin.Context, message string) {
	NotFound(context, "no_such_user", message)
}

// There is no sale with the given ID
// HTTP status code 404
func UnknownSale(context *gin.Context, message string) {
	NotFound(context, "no_such_sale", message)
}

// HTTP status code 403
func WrongUser(context *gin.Context, message string) {
	Forbidden(context, "wrong_user", message)
}

// HTTP status code 404
func UnknownCategory(context *gin.Context, message string) {
	NotFound(context, "no_such_category", message)
}

// HTTP status code 401
func WrongPassword(context *gin.Context, message string) {
	Unauthorized(context, "wrong_password", message)
}

// HTTP status code 403
func CannotUpdateFrozenItem(context *gin.Context, message string) {
	Forbidden(context, "item_frozen", message)
}

// HTTP status code 403
func InvalidPrice(context *gin.Context, message string) {
	Forbidden(context, "invalid_price", message)
}

// HTTP status code 403
func WrongRole(context *gin.Context, message string) {
	Forbidden(context, "wrong_role", message)
}

// HTTP status code 403
func DuplicateItemInSale(context *gin.Context, message string) {
	Forbidden(context, "duplicate_item_in_sale", message)
}

// Seller trying to access other seller's data
// HTTP status code 403
func WrongSeller(context *gin.Context, message string) {
	Forbidden(context, "wrong_seller", message)
}

// HTTP status code 403
func InvalidItemDescription(context *gin.Context, message string) {
	Forbidden(context, "invalid_item_description", message)
}

// HTTP status code 401
func MissingSessionID(context *gin.Context, message string) {
	Unauthorized(context, "missing_session_id", message)
}

// HTTP status code 401
func NoSuchSession(context *gin.Context, message string) {
	Unauthorized(context, "no_such_session", message)
}

// HTTP status code 403
func MissingItems(context *gin.Context, message string) {
	Forbidden(context, "missing_items", message)
}

// HTTP status code 403
func InvalidLayout(context *gin.Context, message string) {
	Forbidden(context, "invalid_layout", message)
}
