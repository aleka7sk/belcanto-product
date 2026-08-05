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
	masterKey     []byte
	digestKey     []byte
	invitationKey []byte
	resetKey      []byte
}

func NewTokenCodec(masterKey []byte) (*TokenCodec, error) {
	if len(masterKey) < 32 {
		return nil, fmt.Errorf("token master key must contain at least 32 bytes")
	}
	held := make([]byte, len(masterKey))
	copy(held, masterKey)
	digestKey := derive(masterKey, "belcanto-token-digest-v1")
	invitationKey := derive(masterKey, "belcanto-invitation-token-v1")
	resetKey := derive(masterKey, "belcanto-password-reset-v1")
	return &TokenCodec{masterKey: held, digestKey: digestKey, invitationKey: invitationKey, resetKey: resetKey}, nil
}

// SecretBox derives an authenticated-encryption box from the same master
// key under a caller-chosen label (e.g. the TOTP secret store).
func (c *TokenCodec) SecretBox(label string) (*SecretBox, error) {
	return NewSecretBox(c.masterKey, label)
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

// PasswordResetToken deterministically derives the recovery token from a
// cryptographically random reset identifier, mirroring InvitationToken:
// PostgreSQL keeps only the digest, and the delivery worker re-derives the
// link from the identifier with the server-held master key.
func (c *TokenCodec) PasswordResetToken(resetID string) string {
	mac := hmac.New(sha256.New, c.resetKey)
	_, _ = mac.Write([]byte("password-reset-id:" + resetID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ContactVerificationCode deterministically derives the six-digit
// confirmation code from a cryptographically random verification
// identifier. Storage keeps only the digest of the code; the delivery
// worker re-derives it from the identifier with the master key.
func (c *TokenCodec) ContactVerificationCode(verificationID string) string {
	mac := hmac.New(sha256.New, c.resetKey)
	_, _ = mac.Write([]byte("contact-verification-id:" + verificationID))
	digest := mac.Sum(nil)
	value := (uint32(digest[0])<<24 | uint32(digest[1])<<16 | uint32(digest[2])<<8 | uint32(digest[3])) % 1_000_000
	return fmt.Sprintf("%06d", value)
}

// NewRecoveryCodes mints one-time 2FA recovery codes in the
// XXXX-XXXX-XX form from an unambiguous alphabet.
func NewRecoveryCodes(count int) ([]string, error) {
	const alphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	codes := make([]string, count)
	for index := range codes {
		raw := make([]byte, 10)
		if _, err := io.ReadFull(rand.Reader, raw); err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}
		letters := make([]byte, 10)
		for position, b := range raw {
			letters[position] = alphabet[int(b)%len(alphabet)]
		}
		codes[index] = string(letters[0:4]) + "-" + string(letters[4:8]) + "-" + string(letters[8:10])
	}
	return codes, nil
}

func EqualDigest(left, right []byte) bool {
	return hmac.Equal(left, right)
}
