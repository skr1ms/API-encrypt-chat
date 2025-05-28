package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"time"
)

// AESEncrypt - шифрует данные с использованием алгоритма AES-256-CBC
func AESEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	start := time.Now()
	defer func() {
		encryptionTime := time.Since(start)
		_ = encryptionTime
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

// AESDecrypt - расшифровывает данные с использованием алгоритма AES-256-CBC
func AESDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	start := time.Now()
	defer func() {
		decryptionTime := time.Since(start)
		_ = decryptionTime
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext)
}

// pkcs7Pad - добавляет дополнение PKCS#7 к данным для выравнивания по блокам
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad - удаляет дополнение PKCS#7 из расшифрованных данных
func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data")
	}

	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.New("invalid padding")
	}

	return data[:(length - unpadding)], nil
}

// GenerateIV - генерирует случайный вектор инициализации для AES шифрования
func GenerateIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	return iv, nil
}
