package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

// RFC 6238 TOTP with the interoperable defaults every authenticator app
// implements: SHA-1, 6 digits, 30-second steps. Verification accepts one
// step of clock skew in each direction.
const totpDigits = 6
const totpStepSeconds = 30
const totpSkewSteps = 1
const totpSecretBytes = 20

var totpEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewTOTPSecret returns a fresh base32 shared secret for enrollment.
func NewTOTPSecret() (string, error) {
	raw := make([]byte, totpSecretBytes)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", fmt.Errorf("generate TOTP secret: %w", err)
	}
	return totpEncoding.EncodeToString(raw), nil
}

// TOTPProvisioningURI renders the otpauth:// URI encoded into the
// enrollment QR code (AUTH-04/ACC-06 step 1).
func TOTPProvisioningURI(issuer, accountLabel, secret string) string {
	label := url.PathEscape(issuer) + ":" + url.PathEscape(accountLabel)
	query := url.Values{}
	query.Set("secret", secret)
	query.Set("issuer", issuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", "6")
	query.Set("period", "30")
	return "otpauth://totp/" + label + "?" + query.Encode()
}

func totpCode(secret string, counter uint64) (string, error) {
	key, err := totpEncoding.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", fmt.Errorf("decode TOTP secret: %w", err)
	}
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(message[:])
	digest := mac.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	code := (binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff) % 1_000_000
	return fmt.Sprintf("%06d", code), nil
}

// TOTPCode renders the current RFC 6238 code for a shared secret at the
// given moment — the same value an authenticator app displays.
func TOTPCode(secret string, at time.Time) (string, error) {
	return totpCode(secret, uint64(at.Unix()/totpStepSeconds))
}

// VerifyTOTP checks a submitted code against the shared secret at the
// given moment, tolerating ±1 time step.
func VerifyTOTP(secret, submitted string, at time.Time) (bool, error) {
	normalized := strings.TrimSpace(submitted)
	if len(normalized) != totpDigits {
		return false, nil
	}
	counter := uint64(at.Unix() / totpStepSeconds)
	for delta := -totpSkewSteps; delta <= totpSkewSteps; delta++ {
		candidate := counter
		if delta < 0 {
			if uint64(-delta) > candidate {
				continue
			}
			candidate -= uint64(-delta)
		} else {
			candidate += uint64(delta)
		}
		expected, err := totpCode(secret, candidate)
		if err != nil {
			return false, err
		}
		if subtle.ConstantTimeCompare([]byte(expected), []byte(normalized)) == 1 {
			return true, nil
		}
	}
	return false, nil
}
