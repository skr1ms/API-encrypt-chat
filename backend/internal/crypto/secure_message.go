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
	MaxTimeDifference = 86400 // 24 часа вместо 5 минут
)

type SecureMessage struct {
	ID             string `json:"id"`
	Timestamp      int64  `json:"timestamp"`
	Nonce          string `json:"nonce"`
	IV             string `json:"iv"`
	Ciphertext     string `json:"ciphertext"`
	HMAC           string `json:"hmac"`
	ECDSASignature string `json:"ecdsa_signature"`
	RSASignature   string `json:"rsa_signature"`
	SenderID       string `json:"sender_id"`
	RecipientID    string `json:"recipient_id"`
}

// CreateSecureMessage - создает зашифрованное сообщение с подписями и целостностью
func CreateSecureMessage(senderID, recipientID string, plaintext []byte, sharedSecret []byte, ecdsaPriv *ecdsa.PrivateKey, rsaPriv *rsa.PrivateKey) (*SecureMessage, error) {

	iv, err := GenerateIV()
	if err != nil {
		return nil, fmt.Errorf("failed to generate IV: %v", err)
	}

	aesKey := sharedSecret[:AESKeySize]

	ciphertext, err := AESEncrypt(aesKey, iv, plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %v", err)
	}

	hmacKey := sharedSecret[AESKeySize : AESKeySize+HMACKeySize]

	hmacValue := GenerateHMAC(hmacKey, ciphertext)

	ecdsaSignature, err := SignECDSA(ecdsaPriv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA signature: %v", err)
	}

	rsaSignature, err := SignRSA(rsaPriv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA signature: %v", err)
	}

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

// VerifyAndDecryptMessage - проверяет целостность и подписи, затем расшифровывает сообщение
func VerifyAndDecryptMessage(msg *SecureMessage, sharedSecret []byte, senderECDSAPublicKey, senderRSAPublicKey []byte) ([]byte, error) {

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

	hmacKey := sharedSecret[AESKeySize : AESKeySize+HMACKeySize]

	if !VerifyHMAC(hmacKey, ciphertext, hmacValue) {
		return nil, errors.New("HMAC verification failed")
	}

	valid, err := VerifyECDSA(senderECDSAPublicKey, ciphertext, ecdsaSignature)
	if err != nil || !valid {
		return nil, fmt.Errorf("ECDSA signature verification failed: %v", err)
	}

	valid, err = VerifyRSA(senderRSAPublicKey, ciphertext, rsaSignature)
	if err != nil || !valid {
		return nil, fmt.Errorf("RSA signature verification failed: %v", err)
	}

	plaintext, err := AESDecrypt(sharedSecret[:AESKeySize], iv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %v", err)
	}

	return plaintext, nil
}

// generateMessageID - генерирует уникальный идентификатор сообщения
func generateMessageID() string {
	nonce, _ := GenerateNonce(16)
	return hex.EncodeToString(nonce)
}
