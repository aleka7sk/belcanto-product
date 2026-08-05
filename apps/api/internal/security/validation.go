package security

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/language"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
var bcp47Pattern = regexp.MustCompile(`^[A-Za-z]{2,8}(?:-[A-Za-z0-9]{1,8})*$`)

func ValidateIdentifier(label, value string, maximumBytes int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maximumBytes || !identifierPattern.MatchString(value) {
		return "", fmt.Errorf("%s must be 1..%d ASCII identifier bytes", label, maximumBytes)
	}
	return value, nil
}

func ValidateText(label, value string, minimumRunes, maximumRunes int) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("%s must be valid UTF-8", label)
	}
	value = strings.TrimSpace(value)
	count := utf8.RuneCountInString(value)
	if count < minimumRunes || count > maximumRunes {
		return "", fmt.Errorf("%s must contain %d..%d Unicode characters", label, minimumRunes, maximumRunes)
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", fmt.Errorf("%s must not contain control characters", label)
		}
	}
	return value, nil
}

func ValidateIdempotencyKey(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 1 || len(value) > 128 {
		return "", fmt.Errorf("Idempotency-Key must contain 1..128 bytes")
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return "", fmt.Errorf("Idempotency-Key must contain visible ASCII characters only")
		}
	}
	return value, nil
}

func ValidateLocale(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 35 {
		return "", fmt.Errorf("locale must contain 1..35 bytes")
	}
	if !bcp47Pattern.MatchString(value) {
		return "", fmt.Errorf("locale must be a valid BCP 47 tag")
	}
	if _, err := language.Parse(value); err != nil {
		return "", fmt.Errorf("locale must be a valid BCP 47 tag")
	}
	return value, nil
}

func ValidateTimezone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 100 {
		return "", fmt.Errorf("timezone must contain 1..100 bytes")
	}
	if _, err := time.LoadLocation(value); err != nil {
		return "", fmt.Errorf("timezone must be a valid IANA location")
	}
	return value, nil
}

// NormalizeEmail lowercases and validates an email address for contact
// verification. The shape check is deliberately conservative: one @, a
// non-empty local part, a dot-separated domain, 254 bytes maximum.
func NormalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if len(normalized) < 3 || len(normalized) > 254 {
		return "", fmt.Errorf("email must contain 3 to 254 characters")
	}
	if !emailPattern.MatchString(normalized) {
		return "", fmt.Errorf("email format is invalid")
	}
	return normalized, nil
}

var emailPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._%+-]*@[a-z0-9][a-z0-9.-]*\.[a-z]{2,}$`)
