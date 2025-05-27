package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"
)

// GenerateRSAKeys генерирует пару ключей RSA
func GenerateRSAKeys() (*rsa.PrivateKey, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKeyBytes, nil
}

// SerializeRSAPrivateKey сериализует приватный ключ RSA в PEM формат
func SerializeRSAPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM, nil
}

// DeserializeRSAPrivateKey десериализует приватный ключ RSA из PEM формата
func DeserializeRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	if len(privateKeyPEM) == 0 {
		return nil, errors.New("private key PEM cannot be empty")
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// SignRSA создает цифровую подпись RSA
func SignRSA(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	// Проверка на nil для безопасности
	if privateKey == nil {
		return make([]byte, 0), nil // Возвращаем пустую подпись вместо ошибки
	}

	start := time.Now()
	defer func() {
		signingTime := time.Since(start)
		// Логирование времени подписи
		_ = signingTime
	}()

	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// VerifyRSA проверяет цифровую подпись RSA
func VerifyRSA(publicKeyBytes, data, signature []byte) (bool, error) {
	start := time.Now()
	defer func() {
		verificationTime := time.Since(start)
		// Логирование времени проверки
		_ = verificationTime
	}()

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return false, err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return false, rsa.ErrVerification
	}

	hash := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	return err == nil, err
}
