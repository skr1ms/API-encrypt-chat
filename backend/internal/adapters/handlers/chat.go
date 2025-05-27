package handlers

import (
	"crypto-chat-backend/internal/crypto"
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/usecase"
	"crypto-chat-backend/internal/infrastructure/websocket"
	"crypto-chat-backend/pkg/logger"
	"crypto/ecdsa"
	"crypto/rsa"
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

	// Преобразуем ответ для отправки расшифрованного контента как основного
	responseMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		responseMessages[i] = map[string]interface{}{
			"id":                msg.Message.ID,
			"chat_id":           msg.Message.ChatID,
			"sender_id":         msg.Message.SenderID,
			"content":           msg.DecryptedContent, // Используем расшифрованный контент как основной
			"decrypted_content": msg.DecryptedContent, // Дублируем для обратной совместимости
			"message_type":      msg.Message.MessageType,
			"created_at":        msg.Message.CreatedAt,
			"updated_at":        msg.Message.UpdatedAt,
			"sender":            msg.Message.Sender,
			"nonce":             msg.Message.Nonce,
			"iv":                msg.Message.IV,
			"hmac":              msg.Message.HMAC,
			"ecdsa_signature":   msg.Message.ECDSASignature,
			"rsa_signature":     msg.Message.RSASignature,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responseMessages,
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
	// Получаем приватные ключи пользователя из базы данных
	// ВНИМАНИЕ: В продакшене приватные ключи должны храниться только на клиенте!
	currentUser := user.(*entities.User)

	var ecdsaPrivateKey *ecdsa.PrivateKey
	var rsaPrivateKey *rsa.PrivateKey

	if currentUser.ECDSAPrivateKey != "" {
		var err error
		ecdsaPrivateKey, err = crypto.DeserializeECDSAPrivateKey([]byte(currentUser.ECDSAPrivateKey))
		if err != nil {
			h.logger.Errorf("Failed to deserialize ECDSA private key: %v", err)
			// Продолжаем с nil ключом
		}
	}

	if currentUser.RSAPrivateKey != "" {
		var err error
		rsaPrivateKey, err = crypto.DeserializeRSAPrivateKey([]byte(currentUser.RSAPrivateKey))
		if err != nil {
			h.logger.Errorf("Failed to deserialize RSA private key: %v", err)
			// Продолжаем с nil ключом
		}
	}

	message, err := h.chatUseCase.SendMessage(uint(chatID), user.(*entities.User).ID, &req, ecdsaPrivateKey, rsaPrivateKey)
	if err != nil {
		h.logger.Errorf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Отправляем сообщение через WebSocket с расшифрованным контентом
	wsMessage := websocket.WSMessage{
		Type:   websocket.MessageTypeChat,
		ChatID: uint(chatID),
		From:   user.(*entities.User).ID,
		Data: websocket.ChatMessage{
			ID:             message.ID,
			ChatID:         message.ChatID,
			SenderID:       message.SenderID,
			Content:        req.Content, // Отправляем оригинальный нешифрованный контент
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

func (h *ChatHandler) CreateOrGetPrivateChat(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		UserID   uint   `json:"user_id" binding:"required"`
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := user.(*entities.User).ID

	// Проверяем, что пользователь не пытается создать чат с самим собой
	if currentUserID == req.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create chat with yourself"})
		return
	}
	chat, err := h.chatUseCase.CreateOrGetPrivateChat(currentUserID, req.UserID, req.Username)
	if err != nil {
		h.logger.Errorf("Failed to create or get private chat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Private chat ready",
		"data":    chat,
	})
}

func (h *ChatHandler) GetChatMembers(c *gin.Context) {
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

	members, err := h.chatUseCase.GetChatMembers(uint(chatID), user.(*entities.User).ID)
	if err != nil {
		h.logger.Errorf("Failed to get chat members: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": members,
	})
}

// SetAdmin назначает пользователя администратором чата
func (h *ChatHandler) SetAdmin(c *gin.Context) {
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
	userIDToUpdate, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.chatUseCase.SetAdmin(uint(chatID), user.(*entities.User).ID, uint(userIDToUpdate))
	if err != nil {
		h.logger.Errorf("Failed to set admin: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User is now an admin"})
}

// RemoveAdmin снимает права администратора с пользователя
func (h *ChatHandler) RemoveAdmin(c *gin.Context) {
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
	userIDToUpdate, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.chatUseCase.RemoveAdmin(uint(chatID), user.(*entities.User).ID, uint(userIDToUpdate))
	if err != nil {
		h.logger.Errorf("Failed to remove admin: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin rights removed"})
}

// LeaveChat позволяет пользователю покинуть групповой чат
func (h *ChatHandler) LeaveChat(c *gin.Context) {
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

	err = h.chatUseCase.LeaveChat(uint(chatID), user.(*entities.User).ID)
	if err != nil {
		h.logger.Errorf("Failed to leave chat: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the chat"})
}

// DeleteChat удаляет приватный чат для пользователя
func (h *ChatHandler) DeleteChat(c *gin.Context) {
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

	err = h.chatUseCase.DeletePrivateChat(uint(chatID), user.(*entities.User).ID)
	if err != nil {
		h.logger.Errorf("Failed to delete chat: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted successfully"})
}

// DeleteGroupChat полностью удаляет групповой чат (только для создателя)
func (h *ChatHandler) DeleteGroupChat(c *gin.Context) {
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

	err = h.chatUseCase.DeleteGroupChat(uint(chatID), user.(*entities.User).ID)
	if err != nil {
		h.logger.Errorf("Failed to delete group chat: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group chat deleted successfully"})
}
