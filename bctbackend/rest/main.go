package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	paths "bctbackend/rest/path"
	"bctbackend/security"
	"database/sql"
	"errors"
	"log/slog"

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
				failure_response.MissingSessionId(context, err.Error())
				return
			}

			sessionData, err := queries.GetSessionData(db, sessionId)

			var noSessionFoundError *queries.NoSessionFoundError
			if errors.As(err, &noSessionFoundError) {
				slog.Info("Session not found")
				failure_response.NoSuchSession(context, err.Error())
				return
			}

			if err != nil {
				slog.Error("Failed to retrieve session from database", slog.String("error", err.Error()))
				failure_response.Unknown(context, "Failed to retrieve session from database: "+err.Error())
				return
			}

			userId := sessionData.UserId
			roleId := sessionData.RoleId

			now := models.Now()
			if err := queries.UpdateLastActivity(db, userId, now); err != nil {
				slog.Error("Failed to update last activity", slog.String("error", err.Error()))
				// Keep going, we don't want to block the request
			}

			handler(context, db, userId, roleId)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST(paths.Login().String(), func(context *gin.Context) { login(context, db) })
	router.POST(paths.Logout().String(), func(context *gin.Context) { logout(context, db) })
	router.GET(paths.Items().String(), withUserAndRole(GetAllItems))
	router.PUT(paths.Items().WithRawItemId(":id"), withUserAndRole(UpdateItem))
	router.GET(paths.Users().String(), withUserAndRole(GetUsers))
	router.GET(paths.Users().WithRawUserId(":id"), withUserAndRole(GetUserInformation))
	router.GET(paths.Categories().String(), withUserAndRole(ListCategories))
	router.GET(paths.SellerItems().WithRawSellerId(":id"), withUserAndRole(GetSellerItems))
	router.POST(paths.SellerItems().WithRawSellerId(":id"), withUserAndRole(AddSellerItem))
	router.POST(paths.Sales().String(), withUserAndRole(AddSale))
	router.GET(paths.Items().WithRawItemId(":id"), withUserAndRole(GetItemInformation))
	router.POST(paths.Labels().WithRawSellerId(":id"), withUserAndRole(GenerateLabels))
	router.GET(paths.Labels().WithRawSellerId(":id"), withUserAndRole(GenerateLabels))
	// router.POST(paths.Labels().WithRawSellerId(":id"), func(context *gin.Context) { GenerateLabels(context, db, 100, models.SellerRoleId) })
}
