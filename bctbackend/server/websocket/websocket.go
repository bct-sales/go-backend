package websocket

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebsocketBroadcaster struct {
	messageChannel chan<- WebsocketBroadcasterMessage
}

type subscriber struct {
	connection *websocket.Conn
	active     bool
	next       *subscriber
}

type broadcasterState struct {
	subscribers    *subscriber
	messageChannel chan<- WebsocketBroadcasterMessage
}

func NewWebsocketBroadcaster() *WebsocketBroadcaster {
	messageChannel := make(chan WebsocketBroadcasterMessage)

	state := broadcasterState{
		subscribers:    nil,
		messageChannel: messageChannel,
	}

	go func() {
		for message := range messageChannel {
			message.execute(&state)
		}
	}()

	return &WebsocketBroadcaster{
		messageChannel: messageChannel,
	}
}

func (wb *WebsocketBroadcaster) AddSubscriber(connection *websocket.Conn) {
	message := &AddSubscriberMessage{
		connection: connection,
	}
	wb.sendMessage(message)
}

func (wb *WebsocketBroadcaster) RemoveSubscriber(subscriber *subscriber) {
	message := &RemoveSubscriberMessage{
		subscriber: subscriber,
	}
	wb.sendMessage(message)
}

func (wb *WebsocketBroadcaster) Broadcast(msg string) {
	message := &BroadcastMessage{
		message: msg,
	}
	wb.sendMessage(message)
}

func (wb *WebsocketBroadcaster) sendMessage(message WebsocketBroadcasterMessage) {
	wb.messageChannel <- message
}

func (wb *WebsocketBroadcaster) CreateHandler() func(*gin.Context) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO Replace with origin check
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
		wb.sendMessage(&message)
	}

	return websocketHandler
}
