package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketServer struct {
	hub *Hub
}

func NewWebSocketServer(hub *Hub) *WebSocketServer {
	return &WebSocketServer{
		hub: hub,
	}
}

func (websocketServer *WebSocketServer) listenAndServe(port string) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			fmt.Printf("Cannot Upgrade to websocket connection: %s \n", err.Error())
			return
		}

		userName := r.URL.Query().Get("userName")
		for client := range websocketServer.hub.clients {
			if userName == client.userName {
				conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Username already taken"))
			}
		}

		client := NewClient(conn, websocketServer.hub, userName)
		websocketServer.hub.register <- client

		go client.readPump()
		go client.writePump()
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error on create server: %s \n", err.Error())
		return
	}
	fmt.Printf("Websocket started on port: %s \n", port)
}
