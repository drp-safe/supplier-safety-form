package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const encryptedPrefix = "enc:"

var encryptionKey []byte

// Init initializes the encryption module with the given key.
// The key must be 32 bytes (AES-256).
func Init(key string) error {
	if key == "" {
		return nil // encryption disabled
	}
	if len(key) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters (AES-256), got %d", len(key))
	}
	encryptionKey = []byte(key)
	return nil
}

// IsEnabled returns true if encryption is configured.
func IsEnabled() bool {
	return len(encryptionKey) > 0
}

// Encrypt encrypts plaintext using AES-256-GCM, returning the value prefixed
// with "enc:". If encryption is disabled the plaintext is returned as-is.
func Encrypt(plaintext string) (string, error) {
	if !IsEnabled() || plaintext == "" {
		return plaintext, nil
	}
	if strings.HasPrefix(plaintext, encryptedPrefix) {
		return plaintext, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return encryptedPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. Values without the "enc:" prefix are returned as-is
// (backwards compatible with plaintext data).
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return ciphertext, nil
	}
	if !strings.HasPrefix(ciphertext, encryptedPrefix) {
		return ciphertext, nil
	}
	if !IsEnabled() {
		return "", errors.New("cannot decrypt: encryption key not configured")
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, encryptedPrefix))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}
