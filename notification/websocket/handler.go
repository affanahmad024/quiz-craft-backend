package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/affanahmad024/quiz-craft/backend/commons"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, check against allowed origins
		return true
	},
}

// ServeWebSocket handles WebSocket requests from clients
func ServeWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get user ID from context (set by auth middleware)
		userIDStr, exists := c.Get("user_id")
		if !exists {
			commons.LogE(ctx, "[websocket.ServeWebSocket::Unauthorized")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			commons.LogE(ctx, "[websocket.ServeWebSocket::InvalidUserID [err=%v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid-user-id"})
			return
		}

		commons.Log(ctx, "[websocket.ServeWebSocket::Upgrading [user_id=%s", userID)

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			commons.LogE(ctx, "[websocket.ServeWebSocket::UpgradeErr [user_id=%s [err=%v", userID, err)
			return
		}

		// Create new client
		client := NewClient(hub, conn, userID)

		// Register client with hub
		hub.register <- client

		commons.Log(ctx, "[websocket.ServeWebSocket::Connected [user_id=%s", userID)

		// Start client pumps in goroutines
		go client.WritePump()
		go client.ReadPump()
	}
}
