package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sleek-chat-backend/internal/crypto"
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/repository"
	"sleek-chat-backend/pkg/logger"
	"time"

	"golang.org/x/crypto/hkdf"
)

type KeyExchangeUseCase struct {
	sessionRepo repository.SessionRepository
	userRepo    repository.UserRepository
	logger      *logger.Logger
}

// NewKeyExchangeUseCase создает новый use case для обмена ключами
func NewKeyExchangeUseCase(
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	logger *logger.Logger,
) *KeyExchangeUseCase {
	return &KeyExchangeUseCase{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// KeyExchangeRequest представляет запрос на обмен ключами
type KeyExchangeRequest struct {
	ClientPublicKey string `json:"clientPublicKey" binding:"required"`
	UserID          uint   `json:"userId" binding:"required"`
}

// KeyExchangeResponse представляет ответ на обмен ключами
type KeyExchangeResponse struct {
	ServerPublicKey string `json:"serverPublicKey"`
	SessionID       string `json:"sessionId"`
	ExpiresAt       int64  `json:"expiresAt"`
}

// SessionInfo содержит информацию о сессии и ключах
type SessionInfo struct {
	SessionID string
	AESKey    []byte
	HMACKey   []byte
	ExpiresAt time.Time
}

// InitiateKeyExchange инициирует процесс обмена ключами с клиентом
func (uc *KeyExchangeUseCase) InitiateKeyExchange(req *KeyExchangeRequest) (*KeyExchangeResponse, *SessionInfo, error) {
	uc.logger.Info("Initiating key exchange", "userID", req.UserID)

	// Проверяем существование пользователя
	user, err := uc.userRepo.GetByID(req.UserID)
	if err != nil {
		uc.logger.Error("User not found", "userID", req.UserID, "error", err)
		return nil, nil, fmt.Errorf("user not found")
	}

	// Генерируем серверную пару ключей ECDH
	serverPrivateKey, serverPublicKeyBytes, err := crypto.GenerateECDSAKeys()
	if err != nil {
		uc.logger.Error("Failed to generate server ECDH keys", "error", err)
		return nil, nil, fmt.Errorf("failed to generate server keys")
	}

	// Декодируем публичный ключ клиента
	clientPublicKeyBytes, err := hex.DecodeString(req.ClientPublicKey)
	if err != nil {
		uc.logger.Error("Failed to decode client public key", "error", err)
		return nil, nil, fmt.Errorf("invalid client public key format")
	}

	// Вычисляем общий секрет ECDH
	sharedSecret, err := crypto.ComputeECDHSharedSecret(serverPrivateKey, clientPublicKeyBytes)
	if err != nil {
		uc.logger.Error("Failed to compute ECDH shared secret", "error", err)
		return nil, nil, fmt.Errorf("failed to compute shared secret")
	}

	// Деривируем AES и HMAC ключи из общего секрета
	aesKey, hmacKey, err := uc.deriveSessionKeys(sharedSecret)
	if err != nil {
		uc.logger.Error("Failed to derive session keys", "error", err)
		return nil, nil, fmt.Errorf("failed to derive session keys")
	}

	// Генерируем уникальный ID сессии
	sessionID, err := uc.generateSessionID()
	if err != nil {
		uc.logger.Error("Failed to generate session ID", "error", err)
		return nil, nil, fmt.Errorf("failed to generate session ID")
	}

	// Создаем сессию в базе данных
	expiresAt := time.Now().Add(24 * time.Hour) // Сессия действительна 24 часа
	session := &entities.Session{
		Token:     sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := uc.sessionRepo.Create(session); err != nil {
		uc.logger.Error("Failed to create session", "error", err)
		return nil, nil, fmt.Errorf("failed to create session")
	}

	// Формируем ответ
	response := &KeyExchangeResponse{
		ServerPublicKey: hex.EncodeToString(serverPublicKeyBytes),
		SessionID:       sessionID,
		ExpiresAt:       expiresAt.Unix(),
	}

	sessionInfo := &SessionInfo{
		SessionID: sessionID,
		AESKey:    aesKey,
		HMACKey:   hmacKey,
		ExpiresAt: expiresAt,
	}

	uc.logger.Info("Key exchange completed successfully",
		"userID", req.UserID,
		"sessionID", sessionID,
		"expiresAt", expiresAt,
	)

	return response, sessionInfo, nil
}

// RefreshSession обновляет существующую сессию и перегенерирует ключи
func (uc *KeyExchangeUseCase) RefreshSession(sessionID string, req *KeyExchangeRequest) (*KeyExchangeResponse, *SessionInfo, error) {
	uc.logger.Info("Refreshing session", "sessionID", sessionID, "userID", req.UserID)

	// Получаем существующую сессию
	session, err := uc.sessionRepo.GetByToken(sessionID)
	if err != nil {
		uc.logger.Error("Session not found", "sessionID", sessionID, "error", err)
		return nil, nil, fmt.Errorf("session not found")
	}

	// Проверяем, что сессия принадлежит пользователю
	if session.UserID != req.UserID {
		uc.logger.Error("Session does not belong to user", "sessionID", sessionID, "userID", req.UserID)
		return nil, nil, fmt.Errorf("unauthorized")
	}

	// Выполняем новый обмен ключами
	return uc.InitiateKeyExchange(req)
}

// ValidateSession проверяет действительность сессии
func (uc *KeyExchangeUseCase) ValidateSession(sessionID string) (*entities.Session, error) {
	session, err := uc.sessionRepo.GetByToken(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	if !session.IsActive {
		return nil, fmt.Errorf("session is inactive")
	}

	if time.Now().After(session.ExpiresAt) {
		// Деактивируем истекшую сессию
		session.IsActive = false
		uc.sessionRepo.Update(session)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// RevokeSession отзывает (деактивирует) сессию
func (uc *KeyExchangeUseCase) RevokeSession(sessionID string) error {
	session, err := uc.sessionRepo.GetByToken(sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	session.IsActive = false
	if err := uc.sessionRepo.Update(session); err != nil {
		uc.logger.Error("Failed to revoke session", "sessionID", sessionID, "error", err)
		return fmt.Errorf("failed to revoke session")
	}

	uc.logger.Info("Session revoked successfully", "sessionID", sessionID)
	return nil
}

// deriveSessionKeys деривирует AES и HMAC ключи из общего секрета
func (uc *KeyExchangeUseCase) deriveSessionKeys(sharedSecret []byte) ([]byte, []byte, error) {
	// Используем HKDF для деривации ключей
	salt := []byte("sleek-chat-salt")
	info := []byte("sleek-chat-session-keys")

	hkdf := hkdf.New(sha256.New, sharedSecret, salt, info)

	// Деривируем 32 байта для AES-256 + 32 байта для HMAC-SHA256
	keys := make([]byte, 64)
	if _, err := hkdf.Read(keys); err != nil {
		return nil, nil, err
	}

	aesKey := keys[:32]  // AES-256 ключ
	hmacKey := keys[32:] // HMAC-SHA256 ключ

	return aesKey, hmacKey, nil
}

// generateSessionID генерирует уникальный ID сессии
func (uc *KeyExchangeUseCase) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
