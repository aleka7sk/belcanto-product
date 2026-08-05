package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// SecretBox provides authenticated encryption at rest for the one secret
// that cannot be digest-only: the TOTP shared secret, which RFC 6238
// verification must read back. The key derives from the server master key
// under a dedicated label, so audit rows and digests never share it.
type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(masterKey []byte, label string) (*SecretBox, error) {
	if len(masterKey) < 32 {
		return nil, fmt.Errorf("secret box master key must contain at least 32 bytes")
	}
	if label == "" {
		return nil, fmt.Errorf("secret box label is required")
	}
	block, err := aes.NewCipher(derive(masterKey, label))
	if err != nil {
		return nil, fmt.Errorf("derive secret box key: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize secret box: %w", err)
	}
	return &SecretBox{aead: aead}, nil
}

func (b *SecretBox) Seal(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate secret box nonce: %w", err)
	}
	return b.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (b *SecretBox) Open(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < b.aead.NonceSize() {
		return nil, fmt.Errorf("secret box ciphertext is truncated")
	}
	nonce := ciphertext[:b.aead.NonceSize()]
	plaintext, err := b.aead.Open(nil, nonce, ciphertext[b.aead.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("open secret box: %w", err)
	}
	return plaintext, nil
}
