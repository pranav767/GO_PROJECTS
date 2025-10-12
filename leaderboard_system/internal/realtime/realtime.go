package realtime

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	broadcast = make(chan []byte, 32) // buffered to reduce drops
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

// Handler upgrades to websocket and registers client
func Handler(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade"})
		return
	}
	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			ws.Close()
			break
		}
	}
}

// Start launches broadcaster goroutine
func Start() {
	go func() {
		for msg := range broadcast {
			clientsMu.Lock()
			for c := range clients {
				_ = c.WriteMessage(websocket.TextMessage, msg)
			}
			clientsMu.Unlock()
		}
	}()
}

// Broadcast sends a message best-effort
func Broadcast(data []byte) {
	select {
	case broadcast <- data:
	default:
	}
}
