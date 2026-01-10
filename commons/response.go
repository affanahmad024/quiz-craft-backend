package commons

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	TidRespHdr = "X-Trace-ID"
)

// APIResponse represents standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RespOK sends 200 OK response
func RespOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// RespOKWithMsg sends 200 OK with message
func RespOKWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

// Resp204 sends 204 No Content
func Resp204(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// RespBadReq sends 400 Bad Request
func RespBadReq(c *gin.Context, errMsg string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: "Bad request",
		Error:   errMsg,
	})
}

// Resp401 sends 401 Unauthorized
func Resp401(c *gin.Context, errMsg string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Message: "Unauthorized",
		Error:   errMsg,
	})
}

// Resp403 sends 403 Forbidden
func Resp403(c *gin.Context, errMsg string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Message: "Forbidden",
		Error:   errMsg,
	})
}

// Resp404 sends 404 Not Found
func Resp404(c *gin.Context, errMsg string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: "Resource not found",
		Error:   errMsg,
	})
}

// Resp409 sends 409 Conflict
func Resp409(c *gin.Context, errMsg string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Message: "Conflict",
		Error:   errMsg,
	})
}

// RespISE sends 500 Internal Server Error
func RespISE(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
		Error:   err.Error(),
	})
}

// RespCustom sends custom status code response
func RespCustom(c *gin.Context, statusCode int, msg string, data interface{}, err string) {
	resp := APIResponse{
		Success: statusCode >= 200 && statusCode < 300,
		Message: msg,
		Data:    data,
		Error:   err,
	}
	c.JSON(statusCode, resp)
}
