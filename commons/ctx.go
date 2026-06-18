package commons

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	TraceIDKey = "trace_id"
	UserIDKey  = "user_id"
)

// Ctx creates a Go context with trace ID from gin context
// If trace ID exists in header (X-Trace-ID), uses it, otherwise generates new one
func Ctx(c *gin.Context) context.Context {
	traceID := c.GetHeader("X-Trace-ID")
	if traceID == "" {
		traceID = uuid.New().String()
	}
	ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)

	// Add user ID to context if available
	if userID, exists := c.Get(UserIDKey); exists {
		ctx = context.WithValue(ctx, UserIDKey, userID)
	}

	return ctx
}

// Tid extracts trace ID from context
func Tid(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return "unknown"
}

// UsrID extracts user ID from context
func UsrID(ctx context.Context) (uuid.UUID, bool) {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID, true
	}
	return uuid.Nil, false
}

// SetUserID sets user ID in gin context
func SetUserID(c *gin.Context, userID uuid.UUID) {
	c.Set(UserIDKey, userID)
	ctx := context.WithValue(c.Request.Context(), UserIDKey, userID)
	c.Request = c.Request.WithContext(ctx)
}
