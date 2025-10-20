package server

import (
	"bctbackend/clock"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/logging"
	"bctbackend/security"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	"bctbackend/server/paths"
	"bctbackend/server/rest"
	"bctbackend/server/websocket"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"embed"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed swagger/ui/*
var swaggerUi embed.FS

type Server struct {
	logger               logging.Logger
	database             *sql.DB
	configuration        *configuration.Configuration
	broadcaster          *websocket.WebsocketBroadcaster
	router               *gin.Engine
	clock                clock.Clock
	expiredSessionTicker clock.Ticker
}

func StartServer(clock clock.Clock, database *sql.DB, logger logging.Logger, configuration *configuration.Configuration) error {
	server, err := NewServer(clock, database, logger, configuration)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer server.Shutdown()

	if err := server.run(); err != nil {
		return err
	}

	return nil
}

func NewServer(clock clock.Clock, db *sql.DB, logger logging.Logger, configuration *configuration.Configuration) (*Server, error) {
	server := Server{
		logger:               logger,
		database:             db,
		configuration:        configuration,
		broadcaster:          websocket.NewWebsocketBroadcaster(),
		router:               createGinRouter(configuration.Server.GinMode, logger),
		clock:                clock,
		expiredSessionTicker: nil,
	}

	server.defineRESTEndpoints()
	server.defineWebsocketEndpoint()
	server.defineStaticFilesRoutes(configuration.Server.HTMLPath)
	server.startPeriodicExpiredSessionPruner()

	return &server, nil
}

func (server *Server) Shutdown() {
	slog.Info("Shutting down server")

	server.expiredSessionTicker.Stop()

	if err := server.database.Close(); err != nil {
		slog.Error("Failed to close database connection", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown complete")
}

func (server *Server) startPeriodicExpiredSessionPruner() {
	clock := server.clock
	pruneInterval := server.configuration.Server.ExpiredSessionPruneInterval

	if server.expiredSessionTicker != nil {
		slog.Error("Expired session ticker already exists, cannot create a new one")
		panic("bug: attempt to create multiple expired session tickers")
	}

	server.expiredSessionTicker = clock.NewTicker(pruneInterval, func() {
		now := clock.Now()
		slog.Info("Cleaning up expired sessions", slog.String("current_time", now.String()))
		if err := queries.DeleteExpiredSessions(server.database, now); err != nil {
			slog.Error("Failed to clean up expired sessions", slog.String("error", err.Error()))
		}
	})
}

func (server *Server) defineRESTEndpoints() {
	router := server.router

	if server.configuration.Server.Swagger {
		slog.Info("Enabling Swagger documentation")
		router.StaticFile("/swagger/swagger.yaml", "./server/swagger/swagger.yaml")

		content, _ := fs.Sub(swaggerUi, "swagger/ui")
		router.StaticFS("/swagger/ui", http.FS(content))
	}

	server.RawPOST(paths.Login(), rest.Login)
	server.RawPOST(paths.Logout(), rest.Logout)

	server.GET(paths.Items(), rest.ListItems)
	server.GET(paths.ItemStr(":id"), rest.GetItemInformation)
	server.PUT(paths.ItemStr(":id"), rest.UpdateItem)

	server.GET(paths.Users(), rest.ListUsers)
	server.GET(paths.UserStr(":id"), rest.GetUserInformation)

	server.GET(paths.Categories(), rest.ListCategories)

	server.GET(paths.SellerItemsStr(":id"), rest.GetSellerItems)
	server.POST(paths.SellerItemsStr(":id"), rest.AddSellerItem)

	server.POST(paths.Labels(), rest.GenerateLabels)

	server.GET(paths.Sales(), rest.GetSales)
	server.GET(paths.SaleStr(":id"), rest.GetSaleInformation)
	server.POST(paths.Sales(), rest.AddSale)
	server.GET(paths.CashierSalesStr(":id"), rest.ListCashierSales)
	server.GET(paths.SoldItems(), rest.ListSoldItems)
}

func (server *Server) defineWebsocketEndpoint() {
	server.router.GET(paths.Websocket().String(), server.broadcaster.CreateHandler())
}

func (server *Server) defineStaticFilesRoutes(htmlPath string) {
	server.router.NoRoute(func(context *gin.Context) {
		context.File(htmlPath)
	})
}

func (server *Server) RawPOST(path *paths.URL, handler func(clock clock.Clock, logger logger.RestLogger, context *gin.Context, database *sql.DB, configuration *configuration.ServerConfiguration)) {
	decoratedSlogger := server.logger.With("handler", getFunctionName(handler))
	logger := logger.NewLoggerWrapper(decoratedSlogger)

	server.router.POST(path.String(), func(context *gin.Context) {
		handler(server.clock, logger, context, server.database, server.configuration.Server)
	})
}

func (server *Server) GET(path *paths.URL, handler rest.HandlerFunction) {
	server.router.GET(path.String(), server.withUserAndRole(handler, false))
}

func (server *Server) POST(path *paths.URL, handler rest.HandlerFunction) {
	server.router.POST(path.String(), server.withUserAndRole(handler, true))
}

func (server *Server) PUT(path *paths.URL, handler rest.HandlerFunction) {
	server.router.PUT(path.String(), server.withUserAndRole(handler, true))
}

func (server *Server) run() error {
	address := fmt.Sprintf("localhost:%d", server.configuration.Server.Port)

	if err := server.router.Run(address); err != nil {
		return err
	}

	return nil
}

func createGinRouter(ginMode string, logger logging.Logger) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		log := logger
		log = log.With("start", time.Now())
		log = log.With("start", time.Now())
		log = log.With("path", c.Request.URL.Path)
		log = log.With("query", c.Request.URL.RawQuery)

		c.Next()

		status := c.Writer.Status()
		log = log.With("status", status)
		log = log.With("method", c.Request.Method)
		log = log.With("host", c.Request.Host)
		log = log.With("route", c.FullPath())
		log = log.With("end", time.Now())
		log = log.With("userAgent", c.Request.UserAgent())
		log = log.With("ip", c.ClientIP())
		log = log.With("referer", c.Request.Referer())

		isError := http.StatusBadRequest <= status
		if isError {
			log.Error("An error occurred while a request was handled")
		} else {
			log.Info("Request successfully handled")
		}
	})

	corsConfiguration := cors.DefaultConfig()
	corsConfiguration.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://localhost:5173"}
	corsConfiguration.AllowCredentials = true
	router.Use(cors.New(corsConfiguration))

	return router
}

func (server *Server) withUserAndRole(handler rest.HandlerFunction, mutates bool) gin.HandlerFunc {
	database := server.database
	configuration := server.configuration
	broadcaster := server.broadcaster
	clock := server.clock

	return func(context *gin.Context) {
		sessionIDString, err := context.Cookie(security.SessionCookieName)
		if err != nil {
			slog.Error("Unauthorized: missing session ID")
			failure_response.MissingSessionID(context, err.Error())
			return
		}

		now := clock.Now()
		sessionID := models.SessionID(sessionIDString)
		sessionData, err := queries.GetSessionData(database, sessionID, now)

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

		userID := sessionData.UserID
		roleID := sessionData.RoleID

		if err := queries.UpdateLastActivity(database, userID, now); err != nil {
			slog.Error("Failed to update last activity", slog.String("error", err.Error()))
			// Keep going, we don't want to block the request
		}

		decoratedSlogger := server.logger.With("user_id", userID.String()).With("role_id", roleID.String()).With("handler", getFunctionName(handler))
		logger := logger.NewLoggerWrapper(decoratedSlogger)
		arguments := rest.HandlerFunctionArguments{
			Clock:         clock,
			Context:       context,
			Configuration: configuration,
			Database:      rest.NewDatabaseWrapper(context, database),
			UserID:        userID,
			RoleID:        roleID,
			Logger:        logger,
		}
		handler(&arguments)

		if mutates {
			broadcaster.Broadcast("update")
		}
	}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if server.router == nil {
		panic("Server router is not initialized")
	}

	server.router.ServeHTTP(writer, request)
}

func getFunctionName(x any) string {
	fullyQualifiedName := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
	indexOfDot := strings.Index(fullyQualifiedName, ".")
	return fullyQualifiedName[indexOfDot+1:]
}
