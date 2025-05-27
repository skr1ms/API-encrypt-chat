package handlers

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/infrastructure/websocket"
	"crypto-chat-backend/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hub    *websocket.Hub
	logger *logger.Logger
}

func NewWebSocketHandler(hub *websocket.Hub, logger *logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	h.hub.ServeWS(c.Writer, c.Request, user.(*entities.User))
}
