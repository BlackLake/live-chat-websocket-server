package main

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	conn     *websocket.Conn
	hub      *Hub
	send     chan []byte
	userName string
}

func NewClient(conn *websocket.Conn, hub *Hub, userName string) *Client {
	return &Client{
		conn:     conn,
		hub:      hub,
		send:     make(chan []byte, 256),
		userName: userName,
	}
}

func (client *Client) readPump() {
	defer func() {
		client.hub.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		client.hub.broadcast <- message
	}
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				client.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := client.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}

	}
}

func (client *Client) write(messageType int, message []byte) error {
	client.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return client.conn.WriteMessage(messageType, message)
}
