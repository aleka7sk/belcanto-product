package security

import (
	"crypto/rand"
	"crypto/subtle"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
	"golang.org/x/text/unicode/norm"
)

const (
	argonMemory       = 19 * 1024
	argonIterations   = 2
	argonParallelism  = 1
	argonSaltLength   = 16
	argonKeyLength    = 32
	minimumRunes      = 15
	maximumRunes      = 128
	maxDenylistItems  = 2_048
	argon2Concurrency = 2
)

//go:embed common_passwords.txt
var embeddedCommonPasswords string

var argon2Slots = make(chan struct{}, argon2Concurrency)
var denylistOnce sync.Once
var embeddedDenylist map[string]struct{}

var ErrInvalidPasswordHash = errors.New("stored password hash is invalid")

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher {
	denylistOnce.Do(func() {
		embeddedDenylist = make(map[string]struct{})
		for _, line := range strings.Split(embeddedCommonPasswords, "\n") {
			value := strings.ToLower(norm.NFC.String(strings.TrimSpace(line)))
			if value == "" || strings.HasPrefix(value, "#") {
				continue
			}
			if len(embeddedDenylist) >= maxDenylistItems {
				panic("embedded password denylist exceeds its bounded maximum")
			}
			embeddedDenylist[value] = struct{}{}
		}
	})
	return &PasswordHasher{}
}

func (h *PasswordHasher) NormalizeAndValidate(password string) (string, error) {
	if !utf8.ValidString(password) {
		return "", fmt.Errorf("password must be valid UTF-8")
	}
	normalized := norm.NFC.String(password)
	runeCount := utf8.RuneCountInString(normalized)
	if runeCount < minimumRunes || runeCount > maximumRunes {
		return "", fmt.Errorf("password must contain between %d and %d Unicode characters", minimumRunes, maximumRunes)
	}
	if _, exists := embeddedDenylist[strings.ToLower(normalized)]; exists {
		return "", fmt.Errorf("password is too common")
	}
	return normalized, nil
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	normalized, err := h.NormalizeAndValidate(password)
	if err != nil {
		return "", err
	}
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	argon2Slots <- struct{}{}
	hash := argon2.IDKey([]byte(normalized), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	<-argon2Slots
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *PasswordHasher) Verify(password, encoded string) bool {
	verified, _ := h.VerifyCredential(password, encoded)
	return verified
}

// VerifyCredential distinguishes an ordinary credential mismatch from a
// malformed stored PHC value. Callers must treat the latter as server-side
// data corruption, not as a user authentication failure.
func (h *PasswordHasher) VerifyCredential(password, encoded string) (bool, error) {
	if !utf8.ValidString(password) {
		return false, nil
	}
	normalized := norm.NFC.String(password)
	runeCount := utf8.RuneCountInString(normalized)
	if runeCount < minimumRunes || runeCount > maximumRunes {
		return false, nil
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, ErrInvalidPasswordHash
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, ErrInvalidPasswordHash
	}
	if parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, iterations, parallelism) ||
		memory != argonMemory || iterations != argonIterations || parallelism != argonParallelism {
		return false, ErrInvalidPasswordHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidPasswordHash
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(salt) < 8 || len(salt) > 64 || len(expected) != argonKeyLength {
		return false, ErrInvalidPasswordHash
	}
	argon2Slots <- struct{}{}
	actual := argon2.IDKey([]byte(normalized), salt, iterations, memory, parallelism, uint32(len(expected)))
	<-argon2Slots
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func (h *PasswordHasher) DummyHash() string {
	// Valid Argon2id PHC string used to equalize the unknown-account path.
	return "$argon2id$v=19$m=19456,t=2,p=1$YmVsY2FudG8tZHVtbXk$Q/GF0d+XAkLFATQJj9G3ydJDU5SeWP6VbkXgYSR1BXw"
}
