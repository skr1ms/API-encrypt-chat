package handlers

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/usecase"
	"crypto-chat-backend/internal/infrastructure/websocket"
	"crypto-chat-backend/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatUseCase *usecase.ChatUseCase
	wsHub       *websocket.Hub
	logger      *logger.Logger
}

func NewChatHandler(chatUseCase *usecase.ChatUseCase, wsHub *websocket.Hub, logger *logger.Logger) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
		wsHub:       wsHub,
		logger:      logger,
	}
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req usecase.CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.chatUseCase.CreateChat(user.(*entities.User).ID, &req)
	if err != nil {
		h.logger.Errorf("Failed to create chat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Chat created successfully",
		"data":    chat,
	})
}

func (h *ChatHandler) GetUserChats(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	chats, err := h.chatUseCase.GetUserChats(user.(*entities.User).ID)
	if err != nil {
		h.logger.Errorf("Failed to get user chats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": chats,
	})
}

func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	messages, err := h.chatUseCase.GetChatMessages(uint(chatID), user.(*entities.User).ID, limit, offset)
	if err != nil {
		h.logger.Errorf("Failed to get chat messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req usecase.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// В реальном приложении здесь нужно получить приватные ключи пользователя
	// Пока что передаем nil (не рекомендуется в продакшене)
	message, err := h.chatUseCase.SendMessage(uint(chatID), user.(*entities.User).ID, &req, nil, nil)
	if err != nil {
		h.logger.Errorf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Отправляем сообщение через WebSocket
	wsMessage := websocket.WSMessage{
		Type:   websocket.MessageTypeChat,
		ChatID: uint(chatID),
		From:   user.(*entities.User).ID,
		Data: websocket.ChatMessage{
			ID:             message.ID,
			ChatID:         message.ChatID,
			SenderID:       message.SenderID,
			Content:        message.Content,
			MessageType:    message.MessageType,
			Nonce:          message.Nonce,
			IV:             message.IV,
			HMAC:           message.HMAC,
			ECDSASignature: message.ECDSASignature,
			RSASignature:   message.RSASignature,
			Timestamp:      message.CreatedAt.Unix(),
		},
	}

	h.wsHub.SendToChat(uint(chatID), wsMessage, user.(*entities.User).ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message sent successfully",
		"data":    message,
	})
}

func (h *ChatHandler) AddMember(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.chatUseCase.AddMember(uint(chatID), user.(*entities.User).ID, req.UserID)
	if err != nil {
		h.logger.Errorf("Failed to add member: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

func (h *ChatHandler) RemoveMember(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	userIDStr := c.Param("userId")
	userIDToRemove, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.chatUseCase.RemoveMember(uint(chatID), user.(*entities.User).ID, uint(userIDToRemove))
	if err != nil {
		h.logger.Errorf("Failed to remove member: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}
