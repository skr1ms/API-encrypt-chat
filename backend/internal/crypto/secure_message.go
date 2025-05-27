package crypto

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	AESKeySize        = 32
	HMACKeySize       = 32
	NonceSize         = 16
	MaxTimeDifference = 300 // секунды (5 минут)
)

// SecureMessage представляет зашифрованное сообщение
type SecureMessage struct {
	ID             string `json:"id"`
	Timestamp      int64  `json:"timestamp"`
	Nonce          string `json:"nonce"`           // hex-encoded
	IV             string `json:"iv"`              // hex-encoded
	Ciphertext     string `json:"ciphertext"`      // hex-encoded
	HMAC           string `json:"hmac"`            // hex-encoded
	ECDSASignature string `json:"ecdsa_signature"` // hex-encoded
	RSASignature   string `json:"rsa_signature"`   // hex-encoded
	SenderID       string `json:"sender_id"`
	RecipientID    string `json:"recipient_id"`
}

// CreateSecureMessage создает защищенное сообщение
func CreateSecureMessage(senderID, recipientID string, plaintext []byte, sharedSecret []byte, ecdsaPriv *ecdsa.PrivateKey, rsaPriv *rsa.PrivateKey) (*SecureMessage, error) {
	// Генерируем IV для AES
	iv, err := GenerateIV()
	if err != nil {
		return nil, fmt.Errorf("failed to generate IV: %v", err)
	}

	// Шифруем сообщение
	ciphertext, err := AESEncrypt(sharedSecret[:AESKeySize], iv, plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %v", err)
	}

	// Создаем HMAC
	hmacValue := GenerateHMAC(sharedSecret[AESKeySize:AESKeySize+HMACKeySize], ciphertext)

	// Создаем подписи
	ecdsaSignature, err := SignECDSA(ecdsaPriv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA signature: %v", err)
	}

	rsaSignature, err := SignRSA(rsaPriv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA signature: %v", err)
	}

	// Генерируем nonce
	nonce, err := GenerateNonce(NonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	timestamp := time.Now().Unix()

	return &SecureMessage{
		ID:             generateMessageID(),
		Timestamp:      timestamp,
		Nonce:          hex.EncodeToString(nonce),
		IV:             hex.EncodeToString(iv),
		Ciphertext:     hex.EncodeToString(ciphertext),
		HMAC:           hex.EncodeToString(hmacValue),
		ECDSASignature: hex.EncodeToString(ecdsaSignature),
		RSASignature:   hex.EncodeToString(rsaSignature),
		SenderID:       senderID,
		RecipientID:    recipientID,
	}, nil
}

// VerifyAndDecryptMessage проверяет и расшифровывает защищенное сообщение
func VerifyAndDecryptMessage(msg *SecureMessage, sharedSecret []byte, senderECDSAPublicKey, senderRSAPublicKey []byte) ([]byte, error) {
	fmt.Printf("DEBUG: VerifyAndDecryptMessage called for message ID: %s\n", msg.ID)

	// Safe ciphertext preview for debug logging
	ciphertextPreview := msg.Ciphertext
	if len(msg.Ciphertext) > 50 {
		ciphertextPreview = msg.Ciphertext[:50] + "..."
	}
	fmt.Printf("DEBUG: Ciphertext: %s\n", ciphertextPreview)

	// Проверяем timestamp
	now := time.Now().Unix()
	if now-msg.Timestamp > MaxTimeDifference || now < msg.Timestamp {
		fmt.Printf("DEBUG: Timestamp check failed - now: %d, msg: %d\n", now, msg.Timestamp)
		return nil, errors.New("timestamp is out of acceptable range")
	}

	// Декодируем hex данные
	ciphertext, err := hex.DecodeString(msg.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %v", err)
	}

	hmacValue, err := hex.DecodeString(msg.HMAC)
	if err != nil {
		return nil, fmt.Errorf("failed to decode HMAC: %v", err)
	}

	ecdsaSignature, err := hex.DecodeString(msg.ECDSASignature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ECDSA signature: %v", err)
	}

	rsaSignature, err := hex.DecodeString(msg.RSASignature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RSA signature: %v", err)
	}

	iv, err := hex.DecodeString(msg.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %v", err)
	}

	// Проверяем HMAC
	if !VerifyHMAC(sharedSecret[AESKeySize:AESKeySize+HMACKeySize], ciphertext, hmacValue) {
		return nil, errors.New("HMAC verification failed")
	}

	// Проверяем ECDSA подпись
	valid, err := VerifyECDSA(senderECDSAPublicKey, ciphertext, ecdsaSignature)
	if err != nil || !valid {
		return nil, fmt.Errorf("ECDSA signature verification failed: %v", err)
	}

	// Проверяем RSA подпись
	valid, err = VerifyRSA(senderRSAPublicKey, ciphertext, rsaSignature)
	if err != nil || !valid {
		return nil, fmt.Errorf("RSA signature verification failed: %v", err)
	}

	// Расшифровываем сообщение
	fmt.Printf("DEBUG: About to decrypt message with AES\n")
	plaintext, err := AESDecrypt(sharedSecret[:AESKeySize], iv, ciphertext)
	if err != nil {
		fmt.Printf("DEBUG: AES decryption failed: %v\n", err)
		return nil, fmt.Errorf("failed to decrypt message: %v", err)
	}

	fmt.Printf("DEBUG: Successfully decrypted message: %s\n", string(plaintext))
	return plaintext, nil
}

func generateMessageID() string {
	nonce, _ := GenerateNonce(16)
	return hex.EncodeToString(nonce)
}
