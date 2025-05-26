package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
)

// GenerateHMAC создает HMAC-SHA256 для данных
func GenerateHMAC(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// VerifyHMAC проверяет HMAC-SHA256 в постоянное время
func VerifyHMAC(key, data, expectedMAC []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	expectedSignature := mac.Sum(nil)

	// Сравнение в постоянное время для защиты от timing attacks
	return subtle.ConstantTimeCompare(expectedMAC, expectedSignature) == 1
}

// GenerateNonce создает случайный nonce
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
