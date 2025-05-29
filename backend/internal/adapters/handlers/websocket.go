package handlers

import (
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/infrastructure/websocket"
	"sleek-chat-backend/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hub    *websocket.Hub
	logger *logger.Logger
}

// NewWebSocketHandler - создает новый экземпляр обработчика WebSocket соединений
func NewWebSocketHandler(hub *websocket.Hub, logger *logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleWebSocket - обрабатывает подключение к WebSocket для авторизованного пользователя
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	h.hub.ServeWS(c.Writer, c.Request, user.(*entities.User))
}
