package handlers

import (
	"sleek-chat-backend/internal/crypto"
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/usecase"
	"sleek-chat-backend/internal/infrastructure/websocket"
	"sleek-chat-backend/pkg/logger"
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

// NewChatHandler - создает новый экземпляр обработчика чатов
func NewChatHandler(chatUseCase *usecase.ChatUseCase, wsHub *websocket.Hub, logger *logger.Logger) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
		wsHub:       wsHub,
		logger:      logger,
	}
}

// CreateChat - создает новый чат
// CreateChat godoc
// @Summary      Create new group chat
// @Description  Creates a new chat with the given name and members
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat  body  map[string]interface{}  true  "Chat name and members"
// @Success      201   {object}  models.Chat
// @Failure      400   {object}  gin.H
// @Router       /chats [post]
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Chat created successfully",
		"data":    chat})
}

// GetUserChats - получает список чатов пользователя
// GetUserChats godoc
// @Summary      Get user chats
// @Description  Returns all chats the authenticated user is a member of
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Chat
// @Router       /chats [get]
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
		"data": chats})
}

// GetChatMessages - получает сообщения чата с постраничной навигацией
// GetChatMessages godoc
// @Summary      Get chat messages
// @Description  Returns all messages from a specific chat
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "Chat ID"
// @Success      200      {array}  models.Message
// @Router       /chats/{chat_id}/messages [get]
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

	responseMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		responseMessages[i] = map[string]interface{}{
			"id":                msg.Message.ID,
			"chat_id":           msg.Message.ChatID,
			"sender_id":         msg.Message.SenderID,
			"content":           msg.DecryptedContent,
			"decrypted_content": msg.DecryptedContent,
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
		"data": responseMessages})
}

// SendMessage - отправляет сообщение в чат с криптографической защитой
// SendMessage godoc
// @Summary      Send message
// @Description  Sends a message to a specific chat
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        message  body  models.Message  true  "Message content"
// @Success      201      {object}  models.Message
// @Failure      400      {object}  gin.H
// @Router       /chats/{chat_id}/messages [get]
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
	currentUser := user.(*entities.User)

	var ecdsaPrivateKey *ecdsa.PrivateKey
	var rsaPrivateKey *rsa.PrivateKey

	if currentUser.ECDSAPrivateKey != "" {
		var err error
		ecdsaPrivateKey, err = crypto.DeserializeECDSAPrivateKey([]byte(currentUser.ECDSAPrivateKey))
		if err != nil {
			h.logger.Errorf("Failed to deserialize ECDSA private key: %v", err)
		}
	}

	if currentUser.RSAPrivateKey != "" {
		var err error
		rsaPrivateKey, err = crypto.DeserializeRSAPrivateKey([]byte(currentUser.RSAPrivateKey))
		if err != nil {
			h.logger.Errorf("Failed to deserialize RSA private key: %v", err)
		}
	}

	message, err := h.chatUseCase.SendMessage(uint(chatID), user.(*entities.User).ID, &req, ecdsaPrivateKey, rsaPrivateKey)
	if err != nil {
		h.logger.Errorf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	wsMessage := websocket.WSMessage{
		Type:   websocket.MessageTypeChat,
		ChatID: uint(chatID),
		From:   user.(*entities.User).ID,
		Data: websocket.ChatMessage{
			ID:             message.ID,
			ChatID:         message.ChatID,
			SenderID:       message.SenderID,
			Content:        req.Content,
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

	responseMessage := map[string]interface{}{
		"id":                message.ID,
		"chat_id":           message.ChatID,
		"sender_id":         message.SenderID,
		"content":           req.Content,
		"decrypted_content": req.Content,
		"message_type":      message.MessageType,
		"created_at":        message.CreatedAt,
		"updated_at":        message.UpdatedAt,
		"sender":            message.Sender,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message sent successfully",
		"data":    responseMessage})
}

// AddMember - добавляет участника в групповой чат
// AddMember godoc
// @Summary      Add member to chat
// @Description  Adds a user to a group chat
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID and username"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id/members [post]
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
	addedUser, err := h.chatUseCase.AddMemberWithUserData(uint(chatID), user.(*entities.User).ID, req.UserID)
	if err != nil {
		h.logger.Errorf("Failed to add member: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member added successfully", "data": gin.H{
			"user": gin.H{
				"id":        addedUser.ID,
				"username":  addedUser.Username,
				"email":     addedUser.Email,
				"is_online": false, // TODO: Get actual online status
				"role":      "member",
			},
		},
	})
}

// RemoveMember - удаляет участника из группового чата
// RemoveMember godoc
// @Summary      Remove member from chat
// @Description  Removes a user from a group chat
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID and username"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id/members/:userId [delete]
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

// CreateOrGetPrivateChat - создает новый приватный чат или возвращает существующий
// CreateOrGetPrivateChat godoc
// @Summary      Create or get private chat
// @Description  Returns existing private chat or creates new one between two users
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body  map[string]string  true  "Username of the other user"
// @Success      200   {object}  models.Chat
// @Failure      400   {object}  gin.H
// @Router       /chats/private [post]
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
		"data":    chat})
}

// GetChatMembers - получает список участников чата
// GetChatMembers godoc
// @Summary      Get chat members
// @Description  Returns a list of members in the chat
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  query  string  true  "Chat ID"
// @Success      200      {array}  string
// @Router       /chats/:id/members [get]
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
		"data": members})
}

// SetAdmin - назначает пользователя администратором чата
// SetAdmin godoc
// @Summary      Set user as admin
// @Description  Promotes a user to admin in the chat
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID and username"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id/members/:userId/admin [put]
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

// RemoveAdmin - снимает административные права с пользователя
// RemoveAdmin godoc
// @Summary      Remove user as admin
// @Description  Demotes a user from admin role in the chat
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID and username"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id/members/:userId/admin [delete]
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

// LeaveChat - позволяет пользователю покинуть чат
// LeaveChat godoc
// @Summary      Leave chat
// @Description  Authenticated user leaves the chat
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id/leave [post]
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

// DeleteChat - удаляет приватный чат
// DeleteChat godoc
// @Summary      Delete chat
// @Description  Deletes chat if user has permission
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id [delete]
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

// DeleteGroupChat - удаляет групповой чат
// DeleteGroupChat godoc
// @Summary      Delete group chat
// @Description  Deletes a group chat and all its messages
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  map[string]string  true  "Chat ID"
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Router       /chats/:id/delete [delete]
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
