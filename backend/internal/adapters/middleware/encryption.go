package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"sleek-chat-backend/internal/crypto"
	"sleek-chat-backend/internal/domain/repository"
	"sleek-chat-backend/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

// EncryptedRequest представляет зашифрованный запрос
type EncryptedRequest struct {
	Data      string `json:"data"`      
	IV        string `json:"iv"`        
	HMAC      string `json:"hmac"`      
	SessionID string `json:"sessionId"` 
}

// EncryptedResponse представляет зашифрованный ответ
type EncryptedResponse struct {
	Data string `json:"data"` 
	IV   string `json:"iv"`  
	HMAC string `json:"hmac"` 
}

// SessionKeys хранит ключи шифрования для сессии
type SessionKeys struct {
	AESKey  []byte
	HMACKey []byte
}

type EncryptionMiddleware struct {
	sessionRepo repository.SessionRepository
	logger      *logger.Logger
	sessionKeys map[string]*SessionKeys
}

// NewEncryptionMiddleware создает новый middleware для шифрования
func NewEncryptionMiddleware(sessionRepo repository.SessionRepository, logger *logger.Logger) *EncryptionMiddleware {
	return &EncryptionMiddleware{
		sessionRepo: sessionRepo,
		logger:      logger,
		sessionKeys: make(map[string]*SessionKeys),
	}
}

// SetSessionKeys устанавливает ключи шифрования для сессии
func (m *EncryptionMiddleware) SetSessionKeys(sessionID string, aesKey, hmacKey []byte) {
	m.sessionKeys[sessionID] = &SessionKeys{
		AESKey:  aesKey,
		HMACKey: hmacKey,
	}
}

// GetSessionKeys получает ключи шифрования для сессии
func (m *EncryptionMiddleware) GetSessionKeys(sessionID string) (*SessionKeys, bool) {
	keys, exists := m.sessionKeys[sessionID]
	return keys, exists
}

// DecryptRequest middleware для расшифровки входящих запросов
func (m *EncryptionMiddleware) DecryptRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			m.logger.Error("Failed to read request body", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		if len(body) == 0 {
			c.Next()
			return
		}

		var encryptedReq EncryptedRequest
		if err := json.Unmarshal(body, &encryptedReq); err != nil {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			c.Next()
			return
		}

		if encryptedReq.Data == "" || encryptedReq.IV == "" || encryptedReq.SessionID == "" {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			c.Next()
			return
		}

		sessionKeys, exists := m.GetSessionKeys(encryptedReq.SessionID)
		if !exists {
			m.logger.Error("Session keys not found", "sessionID", encryptedReq.SessionID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session keys not found"})
			c.Abort()
			return
		}

		encryptedData, err := base64.StdEncoding.DecodeString(encryptedReq.Data)
		if err != nil {
			m.logger.Error("Failed to decode encrypted data", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid encrypted data"})
			c.Abort()
			return
		}

		iv, err := base64.StdEncoding.DecodeString(encryptedReq.IV)
		if err != nil {
			m.logger.Error("Failed to decode IV", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IV"})
			c.Abort()
			return
		}

		if encryptedReq.HMAC != "" {
			providedHMAC, err := base64.StdEncoding.DecodeString(encryptedReq.HMAC)
			if err != nil {
				m.logger.Error("Failed to decode HMAC", "error", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid HMAC"})
				c.Abort()
				return
			}

			calculatedHMAC := crypto.GenerateHMAC(sessionKeys.HMACKey, encryptedData)
			if !crypto.VerifyHMAC(sessionKeys.HMACKey, encryptedData, providedHMAC) {
				m.logger.Error("HMAC verification failed")
				c.JSON(http.StatusBadRequest, gin.H{"error": "HMAC verification failed"})
				c.Abort()
				return
			}

			m.logger.Debug("HMAC verification successful", "calculated", base64.StdEncoding.EncodeToString(calculatedHMAC))
		}

		decryptedData, err := crypto.AESDecrypt(sessionKeys.AESKey, iv, encryptedData)
		if err != nil {
			m.logger.Error("Failed to decrypt request data", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decrypt request data"})
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedData))
		c.Request.ContentLength = int64(len(decryptedData))

		c.Set("sessionID", encryptedReq.SessionID)

		m.logger.Debug("Request decrypted successfully", "sessionID", encryptedReq.SessionID)
		c.Next()
	}
}

// EncryptResponse middleware для шифрования исходящих ответов
func (m *EncryptionMiddleware) EncryptResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		responseWriter := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
			middleware:     m,
			context:        c,
		}
		c.Writer = responseWriter

		c.Next()

		responseWriter.encryptAndWrite()
	}
}

// responseWriterWrapper оборачивает gin.ResponseWriter для перехвата ответа
type responseWriterWrapper struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	middleware *EncryptionMiddleware
	context    *gin.Context
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *responseWriterWrapper) encryptAndWrite() {
	sessionID, exists := w.context.Get("sessionID")
	if !exists {
		w.ResponseWriter.Write(w.body.Bytes())
		return
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		w.ResponseWriter.Write(w.body.Bytes())
		return
	}

	sessionKeys, exists := w.middleware.GetSessionKeys(sessionIDStr)
	if !exists {
		w.ResponseWriter.Write(w.body.Bytes())
		return
	}

	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		w.middleware.logger.Error("Failed to generate IV", "error", err)
		w.ResponseWriter.Write(w.body.Bytes())
		return
	}

	encryptedData, err := crypto.AESEncrypt(sessionKeys.AESKey, iv, w.body.Bytes())
	if err != nil {
		w.middleware.logger.Error("Failed to encrypt response", "error", err)
		w.ResponseWriter.Write(w.body.Bytes())
		return
	}

	hmac := crypto.GenerateHMAC(sessionKeys.HMACKey, encryptedData)

	encryptedResponse := EncryptedResponse{
		Data: base64.StdEncoding.EncodeToString(encryptedData),
		IV:   base64.StdEncoding.EncodeToString(iv),
		HMAC: base64.StdEncoding.EncodeToString(hmac),
	}

	responseData, err := json.Marshal(encryptedResponse)
	if err != nil {
		w.middleware.logger.Error("Failed to marshal encrypted response", "error", err)
		w.ResponseWriter.Write(w.body.Bytes())
		return
	}

	w.ResponseWriter.Header().Set("Content-Type", "application/json")

	w.ResponseWriter.Write(responseData)

	w.middleware.logger.Debug("Response encrypted successfully", "sessionID", sessionIDStr)
}
