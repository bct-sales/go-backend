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

type loginRequest struct {
	Username string `form:"username" binding:"required" json:"username"`
	Password string `form:"password" binding:"required" json:"password"`
}

type loginResponse struct {
	Message string `json:"message"`
}

// @Summary Login user
// @Description Login user
// @Success 200 {object} loginResponse
// @Router /login [post]
// @param username formData string true "username"
// @param password formData string true "password"
func login(context *gin.Context, db *sql.DB) {
	var loginRequest loginRequest

	if err := context.ShouldBind(&loginRequest); err != nil {
		context.String(http.StatusBadRequest, "Bad request: %v", err)
		return
	}

	userId, err := models.ParseId(loginRequest.Username)

	if err != nil {
		slog.Info("Someone tried to login with invalid user ID", slog.String("userId", loginRequest.Username))
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	password := loginRequest.Password

	if err := queries.AuthenticateUser(db, userId, password); err != nil {
		slog.Info("User failed to login", slog.String("userId", loginRequest.Username))
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Login failed"})
		return
	}

	expirationTime := models.Now() + security.SessionDurationInSeconds
	sessionId, err := queries.AddSession(db, userId, expirationTime)

	if err != nil {
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ensureSecure := false // TODO: set to true when using HTTPS
	context.SetCookie(security.SessionCookieName, sessionId, security.SessionDurationInSeconds, "/", "localhost", ensureSecure, true)
	context.JSON(http.StatusOK, loginResponse{Message: "Login successful"})

	slog.Info("User logged in successfully", slog.String("userId", loginRequest.Username))
}
