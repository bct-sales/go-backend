package rest

import (
	"bctbackend/clock"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `binding:"required" form:"username" json:"username"`
	Password string `binding:"required" form:"password" json:"password"`
}

type LoginSuccessResponse struct {
	Role string `json:"role"`
}

// @Summary Login user.
// @Description Login user. If successful, returns the role of the user.
// @Description If the user is unknown, returns 401 Unauthorized with type "unknown_user".
// @Description If the password is wrong, returns 401 Unauthorized with type "wrong_password".
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse request"
// @Failure 401 {object} failure_response.FailureResponse "Failed to authenticate user"
// @Failure 404 {object} failure_response.FailureResponse "Unknown user"
// @Failure 500 {object} failure_response.FailureResponse "Internal error"
// @Router /login [post]
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Tags authentication
func Login(clock clock.Clock, logger logger.RestLogger, context *gin.Context, db *sql.DB, configuration *configuration.ServerConfiguration) {
	var loginRequest LoginRequest

	if err := context.ShouldBind(&loginRequest); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse login request")

		failure_response.InvalidRequest(context, "Failed to parse request")

		return
	}

	userID, err := models.ParseID(loginRequest.Username)
	if err != nil {
		logger.AddInformation("userId", loginRequest.Username)
		logger.InvalidRequest("Someone tried to login with an invalid user ID")

		failure_response.InvalidUserID(context, err.Error())

		return
	}

	password := loginRequest.Password

	logger.AddInformation("userID", userID)
	logger.AddInformation("password", password)

	roleID, err := queries.AuthenticateUser(db, userID, password)

	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			logger.InvalidRequest("Unknown user trying to log in")
			failure_response.UnknownUser(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrWrongPassword) {
			logger.InvalidRequest("User entered wrong password")
			failure_response.WrongPassword(context, err.Error())
			return
		}

		logger.InternalError("Failed authentication for unknown reasons")
		failure_response.Unknown(context, err.Error())
		return
	}

	expirationTime := clock.Now() + security.SessionDurationInSeconds
	sessionID, err := queries.AddSession(db, userID, expirationTime)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to create session")

		failure_response.Unknown(context, err.Error())

		return
	}

	ensureSecure := false // TODO: set to true when using HTTPS
	context.SetCookie(security.SessionCookieName, string(sessionID), security.SessionDurationInSeconds, "/", configuration.CookieDomain, ensureSecure, true)
	roleName := roleID.Name()

	response := LoginSuccessResponse{Role: roleName}
	context.JSON(http.StatusOK, response)
}
