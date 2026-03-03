package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	derivedKey  []byte
	deriveOnce  sync.Once
	deriveError error
)

// DeriveKey reads /etc/machine-id and derives a 32-byte AES-256 key via SHA-256.
func DeriveKey() ([]byte, error) {
	deriveOnce.Do(func() {
		data, err := os.ReadFile("/etc/machine-id")
		if err != nil {
			deriveError = fmt.Errorf("cannot read /etc/machine-id: %w (check permissions)", err)
			return
		}
		hash := sha256.Sum256([]byte(strings.TrimSpace(string(data))))
		derivedKey = hash[:]
	})
	return derivedKey, deriveError
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Returns base64-encoded nonce+ciphertext.
func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decodes base64, splits nonce and ciphertext, and decrypts with AES-256-GCM.
func Decrypt(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
