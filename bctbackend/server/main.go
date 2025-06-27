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
		db:            db,
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
	db            *sql.DB
	configuration *configuration.Configuration
	broadcaster   *websocket.WebsocketBroadcaster
	router        *gin.Engine
}

func (restService *restService) defineEndpoints() {
	DefineEndpoints(restService.db, restService.router, restService.configuration)
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
	db := restService.db
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

func DefineEndpoints(db *sql.DB, router *gin.Engine, configuration *configuration.Configuration) {

	withUserAndRole := func(handler HandlerFunction, mutates bool) gin.HandlerFunc {
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

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST(paths.Login().String(), func(context *gin.Context) { rest.Login(context, db) })
	router.POST(paths.Logout().String(), func(context *gin.Context) { rest.Logout(context, db) })

	router.GET(paths.Items().String(), withUserAndRole(rest.GetAllItems, false))
	router.GET(paths.ItemStr(":id").String(), withUserAndRole(rest.GetItemInformation, false))
	router.PUT(paths.ItemStr(":id").String(), withUserAndRole(rest.UpdateItem, true))

	router.GET(paths.Users().String(), withUserAndRole(rest.GetUsers, false))
	router.GET(paths.UserStr(":id").String(), withUserAndRole(rest.GetUserInformation, false))

	router.GET(paths.Categories().String(), withUserAndRole(rest.ListCategories, false))

	router.GET(paths.SellerItemsStr(":id").String(), withUserAndRole(rest.GetSellerItems, false))
	router.POST(paths.SellerItemsStr(":id").String(), withUserAndRole(rest.AddSellerItem, true))

	router.POST(paths.Labels().String(), withUserAndRole(rest.GenerateLabels, true))

	router.GET(paths.Sales().String(), withUserAndRole(rest.GetSales, false))
	router.GET(paths.SaleStr(":id").String(), withUserAndRole(rest.GetSaleInformation, false))
	router.POST(paths.Sales().String(), withUserAndRole(rest.AddSale, true))
	router.GET(paths.CashierSalesStr(":id").String(), withUserAndRole(rest.GetCashierSales, false))

	router.GET("/api/v1/websocket", broadcaster.CreateHandler())
}
