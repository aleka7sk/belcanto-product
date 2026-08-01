package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type TokenCodec struct {
	digestKey     []byte
	invitationKey []byte
}

func NewTokenCodec(masterKey []byte) (*TokenCodec, error) {
	if len(masterKey) < 32 {
		return nil, fmt.Errorf("token master key must contain at least 32 bytes")
	}
	digestKey := derive(masterKey, "belcanto-token-digest-v1")
	invitationKey := derive(masterKey, "belcanto-invitation-token-v1")
	return &TokenCodec{digestKey: digestKey, invitationKey: invitationKey}, nil
}

func derive(master []byte, label string) []byte {
	mac := hmac.New(sha256.New, master)
	_, _ = mac.Write([]byte(label))
	return mac.Sum(nil)
}

func (c *TokenCodec) NewRawToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (c *TokenCodec) Digest(raw string) []byte {
	mac := hmac.New(sha256.New, c.digestKey)
	_, _ = mac.Write([]byte(raw))
	return mac.Sum(nil)
}

// InvitationToken deterministically derives an opaque, 256-bit token from a
// cryptographically random invitation identifier. The identifier alone cannot
// reproduce the token: the server-held master key is required. This lets an
// idempotent replay return the original link while PostgreSQL stores only the
// token digest, never raw or reversibly encrypted token material.
func (c *TokenCodec) InvitationToken(invitationID string) string {
	mac := hmac.New(sha256.New, c.invitationKey)
	_, _ = mac.Write([]byte("invitation-id:" + invitationID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func EqualDigest(left, right []byte) bool {
	return hmac.Equal(left, right)
}
