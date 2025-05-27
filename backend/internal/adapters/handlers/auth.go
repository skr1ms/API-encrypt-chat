package handlers

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/usecase"
	"crypto-chat-backend/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	logger      *logger.Logger
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req usecase.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Обработка ошибок валидации
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				switch fieldError.Tag() {
				case "required":
					c.JSON(http.StatusBadRequest, gin.H{"error": "MISSING_REQUIRED_FIELD"})
					return
				case "min":
					if fieldError.Field() == "Username" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "USERNAME_TOO_SHORT"})
						return
					}
					if fieldError.Field() == "Password" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "PASSWORD_TOO_SHORT"})
						return
					}
				case "max":
					if fieldError.Field() == "Username" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "USERNAME_TOO_LONG"})
						return
					}
				case "email":
					c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_EMAIL"})
					return
				case "alphanum":
					if fieldError.Field() == "Username" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "USERNAME_INVALID_CHARS"})
						return
					}
				}
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST_DATA"})
		return
	}

	response, err := h.authUseCase.Register(&req)
	if err != nil {
		h.logger.Errorf("Registration failed: %v", err)

		// Определяем статус код на основе типа ошибки
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "USERNAME_ALREADY_EXISTS":
			statusCode = http.StatusConflict
		case "EMAIL_ALREADY_EXISTS":
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data":    response,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST_DATA"})
		return
	}

	response, err := h.authUseCase.Login(&req)
	if err != nil {
		h.logger.Errorf("Login failed: %v", err)

		// Все ошибки логина возвращаем как Unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    response,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token found"})
		return
	}

	err := h.authUseCase.Logout(token.(string))
	if err != nil {
		h.logger.Errorf("Logout failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user.(*entities.User),
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Получаем пользователя из контекста (установлен middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req usecase.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Change password validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userEntity := user.(*entities.User)
	err := h.authUseCase.ChangePassword(userEntity.ID, &req)
	if err != nil {
		h.logger.Errorf("Change password failed: %v", err)

		// Обрабатываем специфичные ошибки
		switch err.Error() {
		case "invalid current password":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current password"})
		case "new password must be different from current password":
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from current password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
