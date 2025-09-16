// Package server implements the WebSocket broadcast server functionality
package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	conn *websocket.Conn
	name string
}

// Convert normal HTTP request to a WebSocket conn.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections in development
	},
}

// Map to store all active clients
var clients = make(map[*Client]bool)

// Use mutex for safe concurrent access
var mu sync.Mutex

// Channel for broadcasting messages
var broadcast = make(chan []byte)

func handleConnection(w http.ResponseWriter, r *http.Request) {
	// Get client name from query parameters
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "anonymous"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
		name: name,
	}
	defer client.conn.Close()

	// After a successful connection add it to the map
	mu.Lock()
	clients[client] = true
	broadcast <- []byte(fmt.Sprintf("System: %s joined the chat", client.name))
	mu.Unlock()

	// Read messages
	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, client)
			broadcast <- []byte(fmt.Sprintf("System: %s left the chat", client.name))
			mu.Unlock()
			break
		}

		if len(msg) > 0 && msg[0] == '/' {
			// Handle commands
			if len(msg) > 6 && string(msg[:6]) == "/name " {
				oldName := client.name
				client.name = string(msg[6:])
				broadcast <- []byte(fmt.Sprintf("System: %s is now known as %s", oldName, client.name))
				continue
			}
		}

		// Broadcast the message with the client's name
		broadcast <- []byte(fmt.Sprintf("%s: %s", client.name, string(msg)))
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		log.Printf("Broadcasting: %s", string(msg))
		mu.Lock()
		for client := range clients {
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error broadcasting to %s: %v", client.name, err)
				client.conn.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

// StartServer starts the WebSocket broadcast server on the specified host and port
func StartServer(host string, port int) {
	http.HandleFunc("/ws", handleConnection)
	go handleMessages()

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Server started at ws://%s/ws\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
