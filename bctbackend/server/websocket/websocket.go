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

func (wun *WebsocketBroadcaster) AddSubscriber(connection *websocket.Conn) {
	message := &AddSubscriberMessage{
		connection: connection,
	}
	wun.sendMessage(message)
}

func (wun *WebsocketBroadcaster) RemoveSubscriber(subscriber *subscriber) {
	message := &RemoveSubscriberMessage{
		subscriber: subscriber,
	}
	wun.sendMessage(message)
}

func (wun *WebsocketBroadcaster) Broadcast(msg string) {
	message := &BroadcastMessage{
		message: msg,
	}
	wun.sendMessage(message)
}

func (wun *WebsocketBroadcaster) sendMessage(message WebsocketBroadcasterMessage) {
	wun.messageChannel <- message
}

func (wun *WebsocketBroadcaster) CreateHandler() func(*gin.Context) {
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
		wun.sendMessage(&message)
	}

	return websocketHandler
}
