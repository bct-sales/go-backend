package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	rest_admin "bctbackend/rest/admin"
	rest_cashier "bctbackend/rest/cashier"
	rest_path "bctbackend/rest/path"
	rest_seller "bctbackend/rest/seller"
	"bctbackend/security"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           BCT Sales
// @version         1.0
// @description     BCT Sales REST API

// @contact.name   Frederic Vogels
// @contact.email  frederic.vogels@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func StartRestService(db *sql.DB) error {
	router := gin.Default()
	SetUpCors(router)
	DefineEndpoints(db, router)

	return router.Run("localhost:8000")
}

func SetUpCors(router *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowCredentials = true

	router.Use(cors.New(config))
}

func DefineEndpoints(db *sql.DB, router *gin.Engine) {
	withUserAndRole := func(handler func(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id)) gin.HandlerFunc {
		return func(context *gin.Context) {
			sessionId, err := context.Cookie(security.SessionCookieName)

			if err != nil {
				slog.Info("Unauthorized: missing session ID")
				context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: missing session ID"})
				return
			}

			sessionData, err := queries.GetSessionData(db, sessionId)

			var noSessionFoundError *queries.NoSessionFoundError
			if errors.As(err, &noSessionFoundError) {
				slog.Info("Session not found")
				context.JSON(http.StatusUnauthorized, gin.H{"message": "Session not found"})
			}

			if err != nil {
				slog.Error("Failed to retrieve session from database", slog.String("error", err.Error()))
				context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve session"})
				return
			}

			userId := sessionData.UserId
			roleId := sessionData.RoleId

			handler(context, db, userId, roleId)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST(rest_path.Login().String(), func(context *gin.Context) { login(context, db) })
	router.POST(rest_path.Logout().String(), func(context *gin.Context) { logout(context, db) })
	router.GET(rest_path.Items().String(), withUserAndRole(rest_admin.GetItems))
	router.GET(rest_path.Users().String(), withUserAndRole(rest_admin.GetUsers))
	router.GET(rest_path.CategoryCounts().String(), withUserAndRole(rest_admin.GetCategoryCounts))
	router.GET(rest_path.SellerItems().WithRawSellerId(":id"), withUserAndRole(rest_seller.GetSellerItems))
	router.POST(rest_path.SellerItems().WithRawSellerId(":id"), withUserAndRole(rest_seller.AddSellerItem))
	router.POST(rest_path.Sales().String(), withUserAndRole(rest_cashier.AddSale))
	router.GET(rest_path.SalesItems().WithRawItemId(":id"), withUserAndRole(rest_cashier.GetItemInformation))
}
