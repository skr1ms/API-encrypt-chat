package handlers

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/usecase"
	"crypto-chat-backend/pkg/logger"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase *usecase.UserUseCase
	logger      *logger.Logger
}

// NewUserHandler - создает новый экземпляр обработчика пользователей
func NewUserHandler(userUseCase *usecase.UserUseCase, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}
}

// SearchUsers - выполняет поиск пользователей по запросу
func (h *UserHandler) SearchUsers(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	h.logger.Info("User found in context", "userType", fmt.Sprintf("%T", user))

	currentUser, ok := user.(*entities.User)
	if !ok {
		h.logger.Error("Invalid user type in context", "userType", fmt.Sprintf("%T", user))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_USER_CONTEXT"})
		return
	}

	userID := currentUser.ID
	h.logger.Info("Processing search request", "userID", userID)

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MISSING_QUERY_PARAMETER"})
		return
	}

	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	req := usecase.SearchUsersRequest{
		Query:  query,
		Limit:  limit,
		UserID: userID,
	}

	result, err := h.userUseCase.SearchUsers(req)
	if err != nil {
		h.logger.Error("Failed to search users", "error", err.Error(), "userID", userID, "query", query)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SEARCH_FAILED"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetUser - получает информацию о пользователе по ID
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_USER_ID"})
		return
	}

	user, err := h.userUseCase.GetUserByID(uint(userID))
	if err != nil {
		h.logger.Error("Failed to get user", "error", err.Error(), "userID", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND"})
		return
	}

	response := gin.H{
		"id":               user.ID,
		"username":         user.Username,
		"email":            user.Email,
		"is_online":        user.IsOnline,
		"last_seen":        user.LastSeen,
		"ecdsa_public_key": user.ECDSAPublicKey,
		"rsa_public_key":   user.RSAPublicKey,
		"created_at":       user.CreatedAt,
	}
	c.JSON(http.StatusOK, response)
}

// GetOnlineUsers - получает список пользователей онлайн
func (h *UserHandler) GetOnlineUsers(c *gin.Context) {
	users, err := h.userUseCase.GetOnlineUsers()
	if err != nil {
		h.logger.Error("Failed to get online users", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FAILED_TO_GET_ONLINE_USERS"})
		return
	}

	response := make([]gin.H, 0, len(users))
	for _, user := range users {
		response = append(response, gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"is_online": user.IsOnline,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"total": len(response),
	})
}
