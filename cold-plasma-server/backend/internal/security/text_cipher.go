package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const encryptedTextPrefix = "enc:v1:"

type TextCipher struct {
	aead cipher.AEAD
}

func NewTextCipher(secret string) (*TextCipher, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("empty encryption secret")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("new aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	return &TextCipher{aead: aead}, nil
}

func (c *TextCipher) Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("random nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, []byte(plain), nil)
	payload := append(nonce, sealed...)
	return encryptedTextPrefix + base64.RawURLEncoding.EncodeToString(payload), nil
}

func (c *TextCipher) DecryptIfEncrypted(value string) string {
	if !strings.HasPrefix(value, encryptedTextPrefix) {
		return value
	}
	raw := strings.TrimPrefix(value, encryptedTextPrefix)
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return "[не удалось расшифровать]"
	}
	nonceSize := c.aead.NonceSize()
	if len(payload) <= nonceSize {
		return "[не удалось расшифровать]"
	}
	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plain, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "[не удалось расшифровать]"
	}
	return string(plain)
}
