package rest

import (
	"database/sql"
	"log/slog"
	"net/http"

	"bctbackend/clock"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	_ "bctbackend/docs"
	"bctbackend/security"
	"bctbackend/server/logger"

	"github.com/gin-gonic/gin"
)

type LogoutPayload struct{}

// @Summary Logout user.
// @Description Logs out the user.
// @Tags authentication
// @Router /logout [post]
func Logout(clock clock.Clock, logger logger.Logger, context *gin.Context, db *sql.DB) {
	sessionIdString, err := context.Cookie(security.SessionCookieName)
	if err != nil {
		logger.InvalidRequest("Cannot logout without session ID", slog.String("error", err.Error()))
		context.JSON(http.StatusOK, gin.H{"message": "Unauthorized: missing session ID"})
		return
	}

	sessionId := models.SessionId(sessionIdString)
	err = queries.DeleteSession(db, sessionId)

	if err != nil {
		logger.InternalError("Failed to delete session", slog.String("error", err.Error()))
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete session"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
