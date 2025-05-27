package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"golang.org/x/crypto/hkdf"
)

// GenerateECDSAKeys генерирует пару ключей ECDSA
func GenerateECDSAKeys() (*ecdsa.PrivateKey, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKeyBytes, nil
}

// SerializeECDSAPrivateKey сериализует приватный ключ ECDSA в PEM формат
func SerializeECDSAPrivateKey(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM, nil
}

// DeserializeECDSAPrivateKey десериализует приватный ключ ECDSA из PEM формата
func DeserializeECDSAPrivateKey(privateKeyPEM []byte) (*ecdsa.PrivateKey, error) {
	if len(privateKeyPEM) == 0 {
		return nil, errors.New("private key PEM cannot be empty")
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// SignECDSA создает цифровую подпись ECDSA
func SignECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	start := time.Now()
	defer func() {
		signingTime := time.Since(start)
		// Логирование времени подписи
		_ = signingTime
	}()

	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// Сериализуем r и s в байты
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifyECDSA проверяет цифровую подпись ECDSA
func VerifyECDSA(publicKeyBytes, data, signature []byte) (bool, error) {
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

	publicKey, ok := publicKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("invalid public key type")
	}

	if len(signature) != 64 { // 32 байта для r + 32 байта для s
		return false, errors.New("invalid signature length")
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	hash := sha256.Sum256(data)
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// ComputeECDHSharedSecret вычисляет общий секрет ECDH
func ComputeECDHSharedSecret(privateKey *ecdsa.PrivateKey, peerPublicKeyBytes []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	if len(peerPublicKeyBytes) == 0 {
		return nil, errors.New("peer public key cannot be empty")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(peerPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := publicKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key type")
	}

	if privateKey.D == nil {
		return nil, errors.New("private key D component is nil")
	}

	x, _ := privateKey.PublicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())
	if x == nil {
		return nil, errors.New("ECDH computation failed")
	}

	// Use HKDF to expand the shared secret to 64 bytes (32 for AES + 32 for HMAC)
	hash := sha256.Sum256(x.Bytes())
	hkdf := hkdf.New(sha256.New, hash[:], nil, []byte("crypto-chat-shared-secret"))

	expandedKey := make([]byte, 64) // 32 bytes for AES + 32 bytes for HMAC
	if _, err := io.ReadFull(hkdf, expandedKey); err != nil {
		return nil, fmt.Errorf("failed to expand shared secret: %v", err)
	}

	return expandedKey, nil
}
