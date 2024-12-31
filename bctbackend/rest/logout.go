package rest

import (
	"database/sql"
	"net/http"

	"bctbackend/database/queries"
	_ "bctbackend/docs"
	"bctbackend/security"

	"github.com/gin-gonic/gin"
)

type LogoutPayload struct{}

// @Summary Logout user.
// @Description Logs out the user.
// @Tags authentication
// @Router /logout [post]
func logout(context *gin.Context, db *sql.DB) {
	sessionId, err := context.Cookie(security.SessionCookieName)

	if err != nil {
		context.JSON(http.StatusOK, gin.H{"message": "Unauthorized: missing session ID"})
		return
	}

	err = queries.DeleteSession(db, sessionId)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete session"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
