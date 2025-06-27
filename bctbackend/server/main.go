package server

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/paths"
	"bctbackend/server/rest"
	"bctbackend/server/websocket"
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
func StartRestService(db *sql.DB, configuration *configuration.Configuration) error {
	restService := restService{
		database:      db,
		configuration: configuration,
		broadcaster:   websocket.NewWebsocketBroadcaster(),
		router:        createGinRouter(),
	}

	restService.defineEndpoints()
	restService.run()

	if err := restService.run(); err != nil {
		return err
	}

	return nil
}

type restService struct {
	database      *sql.DB
	configuration *configuration.Configuration
	broadcaster   *websocket.WebsocketBroadcaster
	router        *gin.Engine
}

func (restService *restService) defineEndpoints() {
	router := restService.router

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	restService.RawPOST(paths.Login(), rest.Login)
	restService.RawPOST(paths.Logout(), rest.Logout)

	restService.GET(paths.Items(), rest.GetAllItems)
	restService.GET(paths.ItemStr(":id"), rest.GetItemInformation)
	restService.PUT(paths.ItemStr(":id"), rest.UpdateItem)

	restService.GET(paths.Users(), rest.GetUsers)
	restService.GET(paths.UserStr(":id"), rest.GetUserInformation)

	restService.GET(paths.Categories(), rest.ListCategories)

	restService.GET(paths.SellerItemsStr(":id"), rest.GetSellerItems)
	restService.POST(paths.SellerItemsStr(":id"), rest.AddSellerItem)

	restService.POST(paths.Labels(), rest.GenerateLabels)

	restService.GET(paths.Sales(), rest.GetSales)
	restService.GET(paths.SaleStr(":id"), rest.GetSaleInformation)
	restService.POST(paths.Sales(), rest.AddSale)
	restService.GET(paths.CashierSalesStr(":id"), rest.GetCashierSales)

	router.GET("/api/v1/websocket", restService.broadcaster.CreateHandler())
}

func (restService *restService) RawPOST(path *paths.URL, handler func(context *gin.Context, database *sql.DB)) {
	restService.router.POST(path.String(), func(context *gin.Context) { handler(context, restService.database) })
}

func (restService *restService) GET(path *paths.URL, handler HandlerFunction) {
	restService.router.GET(path.String(), restService.withUserAndRole(handler, false))
}

func (restService *restService) POST(path *paths.URL, handler HandlerFunction) {
	restService.router.POST(path.String(), restService.withUserAndRole(handler, true))
}

func (restService *restService) PUT(path *paths.URL, handler HandlerFunction) {
	restService.router.PUT(path.String(), restService.withUserAndRole(handler, true))
}

func (restService *restService) run() error {
	if err := restService.router.Run("localhost:8000"); err != nil {
		return err
	}

	return nil
}

func createGinRouter() *gin.Engine {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	return router
}

type HandlerFunction func(context *gin.Context, configuration *configuration.Configuration, db *sql.DB, userId models.Id, roleId models.RoleId)

func (restService *restService) withUserAndRole(handler HandlerFunction, mutates bool) gin.HandlerFunc {
	db := restService.database
	configuration := restService.configuration
	broadcaster := restService.broadcaster

	return func(context *gin.Context) {
		sessionIdString, err := context.Cookie(security.SessionCookieName)
		if err != nil {
			slog.Error("Unauthorized: missing session ID")
			failure_response.MissingSessionId(context, err.Error())
			return
		}

		sessionId := models.SessionId(sessionIdString)
		sessionData, err := queries.GetSessionData(db, sessionId)

		if errors.Is(err, dberr.ErrNoSuchSession) {
			slog.Error("Session not found")
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

		handler(context, configuration, db, userId, roleId)

		if mutates {
			broadcaster.Broadcast("update")
		}
	}
}
