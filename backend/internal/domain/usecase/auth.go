package usecase

import (
	"crypto-chat-backend/internal/crypto"
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtSecret   string
}

func NewAuthUseCase(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
	}
}

type RegisterRequest struct {
	Username       string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	ECDSAPublicKey string `json:"ecdsaPublicKey" binding:"required"`
	RSAPublicKey   string `json:"rsaPublicKey" binding:"required"`
}

type LoginRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	ECDHPublicKey  string `json:"ecdhPublicKey" binding:"required"`
	ECDSAPublicKey string `json:"ecdsaPublicKey" binding:"required"`
	RSAPublicKey   string `json:"rsaPublicKey" binding:"required"`
}

type AuthResponse struct {
	User         *entities.User `json:"user"`
	Token        string         `json:"token"`
	ExpiresAt    time.Time      `json:"expires_at"`
	RefreshToken string         `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

func (uc *AuthUseCase) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Проверяем, существует ли пользователь
	existingUser, _ := uc.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("USERNAME_ALREADY_EXISTS")
	}

	existingUser, _ = uc.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("EMAIL_ALREADY_EXISTS")
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Генерируем ключевые пары
	ecdsaPriv, ecdsaPub, err := crypto.GenerateECDSAKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA keys: %v", err)
	}
	rsaPriv, rsaPub, err := crypto.GenerateRSAKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA keys: %v", err)
	}

	// Сериализуем приватные ключи для сохранения в базе данных
	// ВНИМАНИЕ: Это небезопасно в продакшене! Приватные ключи должны храниться только на клиенте
	ecdsaPrivateKeyPEM, err := crypto.SerializeECDSAPrivateKey(ecdsaPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ECDSA private key: %v", err)
	}

	rsaPrivateKeyPEM, err := crypto.SerializeRSAPrivateKey(rsaPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize RSA private key: %v", err)
	}

	// Создаем пользователя
	user := &entities.User{
		Username:        req.Username,
		Email:           req.Email,
		PasswordHash:    string(hashedPassword),
		ECDSAPublicKey:  hex.EncodeToString(ecdsaPub),
		RSAPublicKey:    hex.EncodeToString(rsaPub),
		ECDSAPrivateKey: string(ecdsaPrivateKeyPEM), // НЕБЕЗОПАСНО в продакшене!
		RSAPrivateKey:   string(rsaPrivateKeyPEM),   // НЕБЕЗОПАСНО в продакшене!
		IsOnline:        false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Создаем токен
	token, expiresAt, err := uc.generateJWT(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Создаем сессию
	session := &entities.Session{
		UserID:       user.ID,
		Token:        token,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	if err := uc.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Сохраняем приватные ключи в возвращаемых данных (в реальном приложении следует передавать их клиенту безопасно)
	_ = ecdsaPriv
	_ = rsaPriv

	return &AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (uc *AuthUseCase) Login(req *LoginRequest) (*AuthResponse, error) {
	// Находим пользователя по username
	user, err := uc.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("INVALID_CREDENTIALS")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("INVALID_CREDENTIALS")
	}

	// Создаем токен
	token, expiresAt, err := uc.generateJWT(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Создаем сессию
	session := &entities.Session{
		UserID:       user.ID,
		Token:        token,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	if err := uc.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Обновляем статус онлайн
	if err := uc.userRepo.UpdateOnlineStatus(user.ID, true); err != nil {
		// Логируем ошибку, но не прерываем процесс
		fmt.Printf("Failed to update online status: %v\n", err)
	}

	return &AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (uc *AuthUseCase) Logout(token string) error {
	session, err := uc.sessionRepo.GetByToken(token)
	if err != nil {
		return err
	}

	// Обновляем статус офлайн
	if err := uc.userRepo.UpdateOnlineStatus(session.UserID, false); err != nil {
		fmt.Printf("Failed to update online status: %v\n", err)
	}

	return uc.sessionRepo.Delete(token)
}

func (uc *AuthUseCase) ValidateToken(tokenString string) (*entities.User, error) {
	// Парсим JWT токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))

		// Проверяем сессию
		session, err := uc.sessionRepo.GetByToken(tokenString)
		if err != nil {
			return nil, errors.New("session not found")
		}

		if session.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("token expired")
		}

		// Обновляем последнюю активность
		uc.sessionRepo.UpdateActivity(tokenString, time.Now())

		return uc.userRepo.GetByID(userID)
	}

	return nil, errors.New("invalid token")
}

func (uc *AuthUseCase) generateJWT(userID uint) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (uc *AuthUseCase) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	// Получаем пользователя
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Проверяем текущий пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("invalid current password")
	}

	// Проверяем, что новый пароль отличается от старого
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.NewPassword)); err == nil {
		return errors.New("new password must be different from current password")
	}

	// Хэшируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Обновляем пароль в базе данных
	if err := uc.userRepo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}
