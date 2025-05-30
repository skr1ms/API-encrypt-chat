package handlers

import (
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/usecase"
	"sleek-chat-backend/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	logger      *logger.Logger
}

// NewAuthHandler - создает новый экземпляр обработчика аутентификации
func NewAuthHandler(authUseCase *usecase.AuthUseCase, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

// Register - обрабатывает запрос на регистрацию нового пользователя
// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user with a username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User credentials"
// @Success      201   {object}  models.User
// @Failure      400   {object}  gin.H
// @Failure      500   {object}  gin.H
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req usecase.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

// Login - обрабатывает запрос на авторизацию пользователя
// Login godoc
// @Summary      Authenticate user
// @Description  Logs in a user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.User  true  "User credentials"
// @Success      200          {object}  map[string]string  "JWT Token"
// @Failure      400          {object}  gin.H
// @Failure      401          {object}  gin.H
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST_DATA"})
		return
	}

	response, err := h.authUseCase.Login(&req)
	if err != nil {
		h.logger.Errorf("Login failed: %v", err)

		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    response,
	})
}

// Logout - обрабатывает запрос на выход пользователя из системы
// Logout godoc
// @Summary      Log out a user
// @Description  Invalidates the user's session or token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  gin.H  "Logged out successfully"
// @Router       /auth/logout [post]
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

// GetProfile - возвращает профиль текущего аутентифицированного пользователя
// @Summary      Get user profile
// @Description  Returns the profile of the authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  gin.H
// @Router       /auth/profile [get]
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

// ChangePassword - обрабатывает запрос на изменение пароля пользователя
// ChangePassword godoc
// @Summary      Change user password
// @Description  Allows the authenticated user to change their password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        passwords  body  map[string]string  true  "Current and new password"
// @Success      200        {object}  gin.H
// @Failure      400        {object}  gin.H
// @Failure      401        {object}  gin.H
// @Failure      500        {object}  gin.H
// @Router       /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
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
