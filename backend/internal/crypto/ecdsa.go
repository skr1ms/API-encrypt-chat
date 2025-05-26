package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"math/big"
	"time"
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

// SignECDSA создает цифровую подпись ECDSA
func SignECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
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
	publicKeyInterface, err := x509.ParsePKIXPublicKey(peerPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := publicKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key type")
	}

	x, _ := privateKey.PublicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())
	hash := sha256.Sum256(x.Bytes())
	return hash[:], nil
}
