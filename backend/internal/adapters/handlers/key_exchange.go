package handlers

import (
	"net/http"
	"sleek-chat-backend/internal/adapters/middleware"
	"sleek-chat-backend/internal/domain/usecase"
	"sleek-chat-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

type KeyExchangeHandler struct {
	keyExchangeUseCase   *usecase.KeyExchangeUseCase
	encryptionMiddleware *middleware.EncryptionMiddleware
	logger               *logger.Logger
}

// NewKeyExchangeHandler создает новый обработчик для обмена ключами
func NewKeyExchangeHandler(
	keyExchangeUseCase *usecase.KeyExchangeUseCase,
	encryptionMiddleware *middleware.EncryptionMiddleware,
	logger *logger.Logger,
) *KeyExchangeHandler {
	return &KeyExchangeHandler{
		keyExchangeUseCase:   keyExchangeUseCase,
		encryptionMiddleware: encryptionMiddleware,
		logger:               logger,
	}
}

// InitiateKeyExchange godoc
// @Summary Initiate key exchange
// @Description Initiates ECDH key exchange with the server and establishes encrypted session
// @Tags key-exchange
// @Accept json
// @Produce json
// @Param request body usecase.KeyExchangeRequest true "Key exchange request"
// @Success 200 {object} usecase.KeyExchangeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/key-exchange/initiate [post]
func (h *KeyExchangeHandler) InitiateKeyExchange(c *gin.Context) {
	var req usecase.KeyExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid key exchange request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.Info("Processing key exchange request", "userID", req.UserID)

	// Выполняем обмен ключами
	response, sessionInfo, err := h.keyExchangeUseCase.InitiateKeyExchange(&req)
	if err != nil {
		h.logger.Error("Key exchange failed", "error", err, "userID", req.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Key exchange failed"})
		return
	}

	// Сохраняем ключи сессии в middleware для будущих запросов
	h.encryptionMiddleware.SetSessionKeys(sessionInfo.SessionID, sessionInfo.AESKey, sessionInfo.HMACKey)

	h.logger.Info("Key exchange successful",
		"userID", req.UserID,
		"sessionID", sessionInfo.SessionID,
	)

	c.JSON(http.StatusOK, response)
}

// RefreshSession godoc
// @Summary Refresh session keys
// @Description Refreshes the encryption keys for an existing session
// @Tags key-exchange
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body usecase.KeyExchangeRequest true "Key exchange request"
// @Success 200 {object} usecase.KeyExchangeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/key-exchange/refresh/{sessionId} [post]
func (h *KeyExchangeHandler) RefreshSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	var req usecase.KeyExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid refresh session request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.Info("Processing session refresh request", "sessionID", sessionID, "userID", req.UserID)

	// Обновляем ключи сессии
	response, sessionInfo, err := h.keyExchangeUseCase.RefreshSession(sessionID, &req)
	if err != nil {
		h.logger.Error("Session refresh failed", "error", err, "sessionID", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session refresh failed"})
		return
	}

	// Обновляем ключи сессии в middleware
	h.encryptionMiddleware.SetSessionKeys(sessionInfo.SessionID, sessionInfo.AESKey, sessionInfo.HMACKey)

	h.logger.Info("Session refresh successful",
		"sessionID", sessionID,
		"newSessionID", sessionInfo.SessionID,
	)

	c.JSON(http.StatusOK, response)
}

// ValidateSession godoc
// @Summary Validate session
// @Description Validates that a session is active and not expired
// @Tags key-exchange
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/key-exchange/validate/{sessionId} [get]
func (h *KeyExchangeHandler) ValidateSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	session, err := h.keyExchangeUseCase.ValidateSession(sessionID)
	if err != nil {
		h.logger.Error("Session validation failed", "error", err, "sessionID", sessionID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session validation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"sessionId": session.Token,
		"userId":    session.UserID,
		"expiresAt": session.ExpiresAt.Unix(),
	})
}

// RevokeSession godoc
// @Summary Revoke session
// @Description Revokes (deactivates) an active session
// @Tags key-exchange
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/key-exchange/revoke/{sessionId} [post]
func (h *KeyExchangeHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	err := h.keyExchangeUseCase.RevokeSession(sessionID)
	if err != nil {
		h.logger.Error("Session revocation failed", "error", err, "sessionID", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session revocation failed"})
		return
	}
	// Удаляем ключи из middleware
	// Нет прямого доступа к методу, создаем новые ключи как nil
	h.encryptionMiddleware.SetSessionKeys(sessionID, nil, nil)

	h.logger.Info("Session revoked successfully", "sessionID", sessionID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session revoked successfully",
	})
}

// GetSessionStatus godoc
// @Summary Get session status
// @Description Returns detailed information about session status
// @Tags key-exchange
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/key-exchange/status/{sessionId} [get]
func (h *KeyExchangeHandler) GetSessionStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	session, err := h.keyExchangeUseCase.ValidateSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid":   false,
			"error":   err.Error(),
			"session": nil,
		})
		return
	}

	// Проверяем наличие ключей в middleware
	_, hasKeys := h.encryptionMiddleware.GetSessionKeys(sessionID)

	c.JSON(http.StatusOK, gin.H{
		"valid":             true,
		"sessionId":         session.Token,
		"userId":            session.UserID,
		"isActive":          session.IsActive,
		"expiresAt":         session.ExpiresAt.Unix(),
		"hasEncryptionKeys": hasKeys,
		"createdAt":         session.CreatedAt.Unix(),
		"updatedAt":         session.UpdatedAt.Unix(),
	})
}

// RegisterRoutes регистрирует маршруты для обмена ключами
func (h *KeyExchangeHandler) RegisterRoutes(router *gin.RouterGroup) {
	keyExchange := router.Group("/key-exchange")
	{
		keyExchange.POST("/initiate", h.InitiateKeyExchange)
		keyExchange.POST("/refresh/:sessionId", h.RefreshSession)
		keyExchange.GET("/validate/:sessionId", h.ValidateSession)
		keyExchange.POST("/revoke/:sessionId", h.RevokeSession)
		keyExchange.GET("/status/:sessionId", h.GetSessionStatus)
	}
}

// RegisterRoutesWithMiddleware регистрирует маршруты с middleware
func (h *KeyExchangeHandler) RegisterRoutesWithMiddleware(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	keyExchange := router.Group("/key-exchange")
	{
		// Публичные маршруты (без аутентификации)
		keyExchange.POST("/initiate", h.InitiateKeyExchange)
		keyExchange.GET("/validate/:sessionId", h.ValidateSession)
		keyExchange.GET("/status/:sessionId", h.GetSessionStatus)

		// Защищенные маршруты (требуют аутентификации)
		protected := keyExchange.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.POST("/refresh/:sessionId", h.RefreshSession)
			protected.POST("/revoke/:sessionId", h.RevokeSession)
		}
	}
}
