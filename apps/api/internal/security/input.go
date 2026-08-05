package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func NormalizePhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	value = replacer.Replace(value)
	if !e164Pattern.MatchString(value) {
		return "", fmt.Errorf("phone must use E.164 format")
	}
	return value, nil
}

func MaskPhone(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
}

func NewID(prefix string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(random), nil
}

func Fingerprint(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode payload fingerprint: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return digest[:], nil
}

// MaskContact renders a privacy-safe preview of a verified contact for
// activation progress and Security Center screens.
func MaskContact(kind, value string) string {
	if kind == "phone" {
		return MaskPhone(value)
	}
	at := strings.IndexByte(value, '@')
	if at <= 0 {
		return "***"
	}
	local, domain := value[:at], value[at+1:]
	visible := 1
	if len(local) > 3 {
		visible = 2
	}
	return local[:visible] + strings.Repeat("*", max(len(local)-visible, 1)) + "@" + domain
}
