package webhook

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
)

func TestDecryptAES_ValidPayload(t *testing.T) {
	plaintext := []byte(`{"data_code":"tisPushData","waybill_number":"YT123"}`)
	encryptKey := "test-key"

	encrypted := encryptForTest(t, plaintext, encryptKey)
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	got, err := DecryptAES(encoded, encryptKey)
	if err != nil {
		t.Fatalf("DecryptAES failed: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("expected %s, got %s", plaintext, got)
	}
}

func TestDecryptAES_EmptyString(t *testing.T) {
	_, err := DecryptAES("", "key")
	if err != ErrEmptyCiphertext {
		t.Fatalf("expected ErrEmptyCiphertext, got: %v", err)
	}
}

func TestDecryptAES_InvalidBase64(t *testing.T) {
	_, err := DecryptAES("not-valid-base64!!!", "key")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidBase64) {
		t.Fatalf("expected ErrInvalidBase64 (possibly wrapped), got: %v", err)
	}
}

func TestDecryptAES_CiphertextTooShort(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("12345678"))
	_, err := DecryptAES(short, "key")
	if err != ErrCiphertextTooShort {
		t.Fatalf("expected ErrCiphertextTooShort, got: %v", err)
	}
}

func TestDecryptAES_WrongKey(t *testing.T) {
	plaintext := []byte(`{"test":true}`)
	encrypted := encryptForTest(t, plaintext, "correct-key")
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	_, err := DecryptAES(encoded, "wrong-key")
	// Wrong key produces garbage with invalid padding
	if err == nil {
		t.Fatal("expected error for wrong key, got nil")
	}
}

func TestDecryptAES_EmptyPlaintext(t *testing.T) {
	// Encrypt empty-ish plaintext (just padding)
	plaintext := []byte(`{}`)
	encryptKey := "test-key"

	encrypted := encryptForTest(t, plaintext, encryptKey)
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	got, err := DecryptAES(encoded, encryptKey)
	if err != nil {
		t.Fatalf("DecryptAES failed: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("expected %s, got %s", plaintext, got)
	}
}

func TestRemovePKCS7Padding_Valid(t *testing.T) {
	// PKCS7: data + padding byte 4,4,4,4
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 4, 4, 4, 4}
	got, err := removePKCS7Padding(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 12 {
		t.Fatalf("expected length 12, got %d", len(got))
	}
}

func TestRemovePKCS7Padding_InvalidPaddingByte(t *testing.T) {
	data := []byte{1, 2, 3, 0} // padding byte 0 is invalid
	_, err := removePKCS7Padding(data)
	if err != ErrInvalidPadding {
		t.Fatalf("expected ErrInvalidPadding, got: %v", err)
	}
}

func TestRemovePKCS7Padding_EmptyData(t *testing.T) {
	_, err := removePKCS7Padding([]byte{})
	if err != ErrInvalidPadding {
		t.Fatalf("expected ErrInvalidPadding for empty data, got: %v", err)
	}
}

// encryptForTest encrypts plaintext with AES-256-CBC for testing.
func encryptForTest(t *testing.T, plaintext []byte, encryptKey string) []byte {
	t.Helper()

	key := sha256.Sum256([]byte(encryptKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatalf("aes.NewCipher failed: %v", err)
	}

	// PKCS7 padding
	padLen := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	// IV + ciphertext
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("rand.Read failed: %v", err)
	}

	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	return append(iv, ciphertext...)
}
