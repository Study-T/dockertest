package webhook

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	aesBlockSize = aes.BlockSize // 16 bytes
	minEncrypted = aesBlockSize * 2
)

var (
	ErrEmptyCiphertext    = errors.New("ciphertext is empty")
	ErrCiphertextTooShort = errors.New("ciphertext too short: need at least 32 bytes (IV + one block)")
	ErrInvalidBase64      = errors.New("invalid base64 in encrypt field")
	ErrInvalidPadding     = errors.New("invalid PKCS#7 padding")
	ErrDecryptionFailed   = errors.New("AES-CBC decryption failed")
)

// DecryptAES decrypts an AES-256-CBC encrypted payload.
// key = SHA256(encryptKey), IV = ciphertext[:16], PKCS#7 padding.
func DecryptAES(base64Ciphertext string, encryptKey string) ([]byte, error) {
	if base64Ciphertext == "" {
		return nil, ErrEmptyCiphertext
	}

	ciphertext, err := base64.StdEncoding.DecodeString(base64Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidBase64, err)
	}

	if len(ciphertext) < minEncrypted {
		return nil, ErrCiphertextTooShort
	}

	key := sha256Sum(encryptKey)
	iv := ciphertext[:aesBlockSize]
	encrypted := ciphertext[aesBlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecryptionFailed, err)
	}

	if len(encrypted)%aesBlockSize != 0 {
		return nil, ErrDecryptionFailed
	}

	plaintext := make([]byte, len(encrypted))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, encrypted)

	plaintext, err = removePKCS7Padding(plaintext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func sha256Sum(data string) []byte {
	h := sha256.Sum256([]byte(data))
	return h[:]
}

func removePKCS7Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidPadding
	}

	padLen := int(data[len(data)-1])
	if padLen < 1 || padLen > aesBlockSize || padLen > len(data) {
		return nil, ErrInvalidPadding
	}

	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, ErrInvalidPadding
		}
	}

	return data[:len(data)-padLen], nil
}
