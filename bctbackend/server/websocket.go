package server

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

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
