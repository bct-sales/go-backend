package server

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	paths "bctbackend/server/path"
	"bctbackend/security"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	_ "bctbackend/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
func StartRestService(db *sql.DB, configuration *Configuration) error {
	router := gin.Default()
	SetUpCors(router)
	DefineEndpoints(db, router, configuration)

	return router.Run("localhost:8000")
}

func SetUpCors(router *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowCredentials = true

	router.Use(cors.New(config))
}

type HandlerFunction func(context *gin.Context, configuration *Configuration, db *sql.DB, userId models.Id, roleId models.RoleId)

func DefineEndpoints(db *sql.DB, router *gin.Engine, configuration *Configuration) {
	broadcasterChannel := NewWebsocketBroadcaster()

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
				message := &BroadcastMessage{message: "update"}
				broadcasterChannel <- message
			}
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST(paths.Login().String(), func(context *gin.Context) { login(context, db) })
	router.POST(paths.Logout().String(), func(context *gin.Context) { logout(context, db) })

	router.GET(paths.Items().String(), withUserAndRole(GetAllItems, false))
	router.GET(paths.Items().IdStr(":id").String(), withUserAndRole(GetItemInformation, false))
	router.PUT(paths.Items().IdStr(":id").String(), withUserAndRole(UpdateItem, true))

	router.GET(paths.Users().String(), withUserAndRole(GetUsers, false))
	router.GET(paths.Users().WithRawUserId(":id"), withUserAndRole(GetUserInformation, false))

	router.GET(paths.Categories().String(), withUserAndRole(ListCategories, false))

	router.GET(paths.SellerItems().WithRawSellerId(":id"), withUserAndRole(GetSellerItems, false))
	router.POST(paths.SellerItems().WithRawSellerId(":id"), withUserAndRole(AddSellerItem, true))

	router.POST(paths.Labels().String(), withUserAndRole(GenerateLabels, true))

	router.GET(paths.Sales().String(), withUserAndRole(GetSales, false))
	router.GET(paths.Sales().IdStr(":id").String(), withUserAndRole(GetSaleInformation, false))
	router.POST(paths.Sales().String(), withUserAndRole(AddSale, true))
	router.GET(paths.CashierSales().WithRawCashierId(":id"), withUserAndRole(GetCashierSales, false))

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Replace with origin check in production
		},
	}

	websocketHandler := func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Error("Failed to upgrade connection to WebSocket", slog.String("error", err.Error()))
			return
		}

		message := AddSubscriberMessage{
			connection: conn,
		}
		broadcasterChannel <- &message
	}

	router.GET("/api/v1/websocket", websocketHandler)
}

type WebsocketBroadcasterMessage interface {
	execute(broadcasterState *BroadcasterState)
}

type AddSubscriberMessage struct {
	connection *websocket.Conn
}

func (m *AddSubscriberMessage) execute(broadcasterState *BroadcasterState) {
	subscriber := &Subscriber{
		connection: m.connection,
		active:     true,
		next:       broadcasterState.subscribers,
	}

	broadcasterState.subscribers = subscriber

	go func() {
		m.connection.SetReadDeadline(time.Time{})

		_, _, err := m.connection.ReadMessage()
		if err != nil {
			slog.Error("Failed to read message from WebSocket", slog.String("error", err.Error()))
		}

		broadcasterState.messageChannel <- &RemoveSubscriberMessage{subscriber: subscriber}
	}()
}

type RemoveSubscriberMessage struct {
	subscriber *Subscriber
}

func (m *RemoveSubscriberMessage) execute(broadcasterState *BroadcasterState) {
	m.subscriber.active = false
	m.subscriber.connection.Close()
}

type BroadcastMessage struct {
	message string
}

func (m *BroadcastMessage) execute(broadcasterState *BroadcasterState) {
	var dummy *Subscriber
	previousSubscriberNext := &dummy

	current := broadcasterState.subscribers

	for current != nil {
		if !current.active {
			*previousSubscriberNext = current.next
		} else {
			current.connection.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := current.connection.WriteMessage(websocket.TextMessage, []byte(m.message)); err != nil {
				slog.Error("Failed to write message to WebSocket", slog.String("error", err.Error()))
				current.active = false
				current.connection.Close()
			}

			previousSubscriberNext = &current.next
		}

		current = current.next
	}
}

type Subscriber struct {
	connection *websocket.Conn
	active     bool
	next       *Subscriber
}

type BroadcasterState struct {
	subscribers    *Subscriber
	messageChannel chan<- WebsocketBroadcasterMessage
}

func NewWebsocketBroadcaster() chan<- WebsocketBroadcasterMessage {
	messageChannel := make(chan WebsocketBroadcasterMessage)

	state := BroadcasterState{
		subscribers:    nil,
		messageChannel: messageChannel,
	}

	go func() {
		for message := range messageChannel {
			message.execute(&state)
		}
	}()

	return messageChannel
}
