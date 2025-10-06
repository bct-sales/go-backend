package rest

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"bctbackend/clock"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"

	"github.com/gin-gonic/gin"
)

type LogoutPayload struct{}

// @Summary Logout user.
// @Description Logs out the user.
// @Tags authentication
// @Router /logout [post]
func Logout(clock clock.Clock, logger logger.Logger, context *gin.Context, db *sql.DB, configuration *configuration.ServerConfiguration) {
	sessionIDString, err := context.Cookie(security.SessionCookieName)
	if err != nil {
		logger.InvalidRequest("Cannot logout without session ID", slog.String("error", err.Error()))
		context.JSON(http.StatusOK, gin.H{"message": "Unauthorized: missing session ID"})
		return
	}

	sessionID := models.SessionID(sessionIDString)
	err = queries.DeleteSession(db, sessionID)

	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSession) {
			failure_response.NoSuchSession(context, "No such session - perhaps it has expired and was pruned")
			return
		}

		logger.InternalError("Failed to delete session", slog.String("error", err.Error()))
		failure_response.Unknown(context, "Failed to delete session")
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
