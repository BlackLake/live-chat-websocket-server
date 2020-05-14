package main

func main() {
	hub := NewHub()
	go hub.Run()

	webSocketServer := NewWebSocketServer(hub)

	webSocketServer.listenAndServe("8080")
}
