package ws

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket client connections and message broadcasting.
// It implements both http.Handler (for serving WebSocket upgrades)
// and service.Broadcaster (for sending messages to all clients).
type Hub struct {
	clients   map[*websocket.Conn]bool
	mu        sync.Mutex
	broadcast chan []byte
	upgrader  websocket.Upgrader
	logger    *slog.Logger
	done      chan struct{}
}

// NewHub creates a new WebSocket Hub.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte, 32),
		upgrader:  websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		logger:    logger,
		done:      make(chan struct{}),
	}
}

// Start launches the broadcaster goroutine.
func (h *Hub) Start() {
	go func() {
		for {
			select {
			case msg, ok := <-h.broadcast:
				if !ok {
					return
				}
				h.mu.Lock()
				for c := range h.clients {
					_ = c.WriteMessage(websocket.TextMessage, msg)
				}
				h.mu.Unlock()
			case <-h.done:
				return
			}
		}
	}()
}

// Broadcast sends a message to all connected clients.
// Implements service.Broadcaster interface.
func (h *Hub) Broadcast(data []byte) {
	select {
	case h.broadcast <- data:
	default:
	}
}

// ServeHTTP handles WebSocket upgrade requests.
// Implements http.Handler interface.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", slog.Any("error", err))
		return
	}
	h.logger.Info("websocket client connected")

	h.mu.Lock()
	h.clients[ws] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, ws)
		h.mu.Unlock()
		ws.Close()
	}()

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			h.logger.Info("websocket client disconnected", slog.Any("error", err))
			break
		}
	}
}

// Shutdown gracefully drains all WebSocket connections.
func (h *Hub) Shutdown() {
	close(h.done)
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.Close()
		delete(h.clients, c)
	}
}
