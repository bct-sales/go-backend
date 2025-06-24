package websocket

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketBroadcasterMessage interface {
	execute(broadcasterState *broadcasterState)
}

type AddSubscriberMessage struct {
	connection *websocket.Conn
}

func (m *AddSubscriberMessage) execute(broadcasterState *broadcasterState) {
	subscriber := &subscriber{
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
	subscriber *subscriber
}

func (m *RemoveSubscriberMessage) execute(broadcasterState *broadcasterState) {
	m.subscriber.active = false
	m.subscriber.connection.Close()
}

type BroadcastMessage struct {
	message string
}

func (m *BroadcastMessage) execute(broadcasterState *broadcasterState) {
	var dummy *subscriber
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
