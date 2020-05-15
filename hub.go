package main

import (
	"encoding/json"
	"fmt"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.registerClient(client)
			break
		case client := <-hub.unregister:
			hub.unregisterClient(client)
			break
		case message := <-hub.broadcast:
			hub.broadcastMessage(message)
		}
	}
}

func (hub *Hub) registerClient(client *Client) {
	hub.clients[client] = true

	messageBytes, err := json.Marshal(Message{Sender: client.userName, Message: "Connected"})
	if err == nil {
		hub.broadcastMessage(messageBytes)
	}

	fmt.Printf("Client connected from : %s \n", client.conn.RemoteAddr())
}

func (hub *Hub) unregisterClient(client *Client) {
	_, ok := hub.clients[client]
	if ok {
		delete(hub.clients, client)
		close(client.send)
	}

	messageBytes, err := json.Marshal(Message{Sender: client.userName, Message: "Disconnected"})
	if err == nil {
		hub.broadcastMessage(messageBytes)
	}
}

func (hub *Hub) broadcastMessage(message []byte) {
	for client := range hub.clients {
		go func(client *Client) {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(hub.clients, client)
			}
		}(client)
	}
}
