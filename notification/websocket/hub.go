package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/notification/model"
)

// Hub maintains active WebSocket connections and broadcasts messages
type Hub struct {
	// Registered clients by user ID
	clients map[uuid.UUID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific user
	broadcast chan *UserMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// UserMessage represents a message to be sent to a specific user
type UserMessage struct {
	UserID       uuid.UUID
	Notification model.NotificationAPI
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *UserMessage, 256),
		mu:         sync.RWMutex{},
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToUser(message)
		}
	}
}

// registerClient registers a new client connection
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true
}

// unregisterClient unregisters a client connection
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			// Remove user entry if no more clients
			if len(clients) == 0 {
				delete(h.clients, client.UserID)
			}
		}
	}
}

// broadcastToUser sends a message to all connections of a specific user
func (h *Hub) broadcastToUser(message *UserMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[message.UserID]; ok {
		messageBytes, err := json.Marshal(message.Notification)
		if err != nil {
			return
		}

		for client := range clients {
			select {
			case client.send <- messageBytes:
			default:
				// Client's send channel is full, skip
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

// BroadcastToUser broadcasts a notification to a specific user (implements WebSocketHub interface)
func (h *Hub) BroadcastToUser(userID uuid.UUID, notification model.NotificationAPI) {
	h.broadcast <- &UserMessage{
		UserID:       userID,
		Notification: notification,
	}
}

// GetConnectedUsersCount returns the number of users with active connections
func (h *Hub) GetConnectedUsersCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetUserConnectionsCount returns the number of active connections for a specific user
func (h *Hub) GetUserConnectionsCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}
