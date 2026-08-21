package net

import (
	"github.com/gorilla/websocket"
)

type SocketEvent string

const (
	SOCKET_EVENT_ACK        SocketEvent = "ack"
	SOCKET_EVENT_CONNECT    SocketEvent = "connect"
	SOCKET_EVENT_DISCONNECT SocketEvent = "disconnect"
	SOCKET_EVENT_ERROR      SocketEvent = "error"
)

type SocketMessage struct {
	Event   SocketEvent `json:"event"`
	Payload any         `json:"data"`
}

func BroadcastSocketMessage(conns []*websocket.Conn, msg *SocketMessage) []string {
	var errorMessages []string
	for _, conn := range conns {
		err := conn.WriteJSON(msg)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}
	if len(errorMessages) == 0 {
		return nil
	}
	return errorMessages
}

func NewSocketMessage(event SocketEvent, payload any) *SocketMessage {
	return &SocketMessage{
		Event:   event,
		Payload: payload,
	}
}

func SendSocketMessage(conn *websocket.Conn, msg *SocketMessage) error {
	err := conn.WriteJSON(msg)
	return err
}
