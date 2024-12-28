package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"database/sql"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `form:"username" binding:"required" json:"username"`
	Password string `form:"password" binding:"required" json:"password"`
}

type LoginSuccessResponse struct {
	Role string `json:"role"`
}

type LoginFailureResponse struct {
	Message string `json:"message"`
}

// @Summary Login user
// @Description Login user
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} LoginFailureResponse "Failed to parse request"
// @Failure 401 {object} LoginFailureResponse "Failed to authenticate user"
// @Failure 500 {object} LoginFailureResponse "Internal error"
// @Router /login [post]
// @param username formData string true "username"
// @param password formData string true "password"
func login(context *gin.Context, db *sql.DB) {
	var loginRequest LoginRequest

	if err := context.ShouldBind(&loginRequest); err != nil {
		slog.Info("Failed to parse login request", slog.String("error", err.Error()))
		failureResponse := LoginFailureResponse{Message: "Bad request: " + err.Error()}
		context.JSON(http.StatusBadRequest, failureResponse)
		return
	}

	userId, err := models.ParseId(loginRequest.Username)

	if err != nil {
		slog.Info("Someone tried to login with invalid user ID", slog.String("userId", loginRequest.Username))
		failureResponse := LoginFailureResponse{Message: "Invalid user ID"}
		context.JSON(http.StatusUnauthorized, failureResponse)
		return
	}

	password := loginRequest.Password
	roleId, err := queries.AuthenticateUser(db, userId, password)

	if err != nil {
		slog.Info("Failed to authenticate user", slog.String("userId", loginRequest.Username))
		failureResponse := LoginFailureResponse{Message: "Failed to authenticate user"}
		context.JSON(http.StatusUnauthorized, failureResponse)
		return
	}

	expirationTime := models.Now() + security.SessionDurationInSeconds
	sessionId, err := queries.AddSession(db, userId, expirationTime)

	if err != nil {
		failureResponse := LoginFailureResponse{Message: "An error occurred while creating a session"}
		context.AbortWithStatusJSON(http.StatusInternalServerError, failureResponse)
		return
	}

	ensureSecure := false // TODO: set to true when using HTTPS
	context.SetCookie(security.SessionCookieName, sessionId, security.SessionDurationInSeconds, "/", "localhost", ensureSecure, true)
	roleName, err := models.NameOfRole(roleId)

	if err != nil {
		slog.Error("Failed to get role name", slog.Int64("roleId", int64(roleId)))
		failureResponse := LoginFailureResponse{Message: "User has an unknown role"}
		context.JSON(http.StatusInternalServerError, failureResponse)
		return
	}

	response := LoginSuccessResponse{Role: roleName}
	context.JSON(http.StatusOK, response)

	slog.Info("User logged in successfully", slog.String("userId", loginRequest.Username))
}
