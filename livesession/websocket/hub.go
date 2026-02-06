package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Hub maintains WebSocket connections for live sessions
type Hub struct {
	// Registered clients by session ID
	sessions map[uuid.UUID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific session
	broadcast chan *SessionMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// SessionMessage represents a message to be sent to a session
type SessionMessage struct {
	SessionID uuid.UUID
	Event     Event
}

// Event represents a live session event
type Event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		sessions:   make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *SessionMessage, 256),
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
			h.broadcastToSession(message)
		}
	}
}

// registerClient registers a new client connection
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sessions[client.SessionID] == nil {
		h.sessions[client.SessionID] = make(map[*Client]bool)
	}
	h.sessions[client.SessionID][client] = true
}

// unregisterClient unregisters a client connection
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.sessions[client.SessionID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			// Remove session entry if no more clients
			if len(clients) == 0 {
				delete(h.sessions, client.SessionID)
			}
		}
	}
}

// broadcastToSession sends a message to all connections in a session
func (h *Hub) broadcastToSession(message *SessionMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.sessions[message.SessionID]; ok {
		messageBytes, err := json.Marshal(message.Event)
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

// BroadcastToSession broadcasts an event to a specific session
func (h *Hub) BroadcastToSession(sessionID uuid.UUID, event Event) {
	h.broadcast <- &SessionMessage{
		SessionID: sessionID,
		Event:     event,
	}
}

// GetSessionConnectionCount returns the number of active connections for a session
func (h *Hub) GetSessionConnectionCount(sessionID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.sessions[sessionID]; ok {
		return len(clients)
	}
	return 0
}

// GetTotalConnectionsCount returns the total number of connections across all sessions
func (h *Hub) GetTotalConnectionsCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, clients := range h.sessions {
		total += len(clients)
	}
	return total
}

// Event type constants
const (
	EventSessionStarted        = "session_started"
	EventQuestionRevealed      = "question_revealed"
	EventParticipantJoined     = "participant_joined"
	EventParticipantLeft       = "participant_left"
	EventAnswerSubmitted       = "answer_submitted"
	EventLeaderboardUpdate     = "leaderboard_update"
	EventQuestionEnded         = "question_ended"
	EventSessionEnded          = "session_ended"
	EventTimerUpdate           = "timer_update"
	EventWholeQuizTimerUpdate  = "whole_quiz_timer_update"
	EventQuestionAutoAdvance   = "question_auto_advance"
)
