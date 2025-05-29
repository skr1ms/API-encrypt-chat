package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
)

// GenerateHMAC - создает HMAC-SHA256 хеш для переданных данных
func GenerateHMAC(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	result := mac.Sum(nil)

	return result
}

// VerifyHMAC - проверяет соответствие HMAC-SHA256 в постоянное время
func VerifyHMAC(key, data, expectedMAC []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	expectedSignature := mac.Sum(nil)

	result := subtle.ConstantTimeCompare(expectedMAC, expectedSignature) == 1

	return result
}

// GenerateNonce - создает случайный nonce указанного размера
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
