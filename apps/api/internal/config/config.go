package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment             string
	DatabaseURL             string
	ListenAddress           string
	TokenMasterKey          []byte
	PublicActivationBaseURL string
	AccessTTL               time.Duration
	RefreshTTL              time.Duration
	InvitationTTL           time.Duration
	AutoMigrate             bool
	MediaDir                string
}

func Load() (Config, error) {
	environment := strings.TrimSpace(envOr("APP_ENV", "development"))
	switch environment {
	case "development", "test", "production":
	default:
		return Config{}, fmt.Errorf("APP_ENV must be development, test, or production")
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	keyText := os.Getenv("TOKEN_HMAC_KEY_BASE64")
	key, err := base64.StdEncoding.DecodeString(keyText)
	if err != nil || len(key) < 32 {
		return Config{}, fmt.Errorf("TOKEN_HMAC_KEY_BASE64 must decode to at least 32 bytes")
	}

	autoMigrate, err := strconv.ParseBool(envOr("AUTO_MIGRATE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTO_MIGRATE: %w", err)
	}
	activationBaseURL := strings.TrimSpace(os.Getenv("PUBLIC_ACTIVATION_BASE_URL"))
	if activationBaseURL == "" {
		return Config{}, fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL is required")
	}
	activationBaseURL, err = validateActivationBaseURL(activationBaseURL, environment)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:             environment,
		DatabaseURL:             databaseURL,
		ListenAddress:           envOr("LISTEN_ADDRESS", ":8080"),
		TokenMasterKey:          key,
		PublicActivationBaseURL: activationBaseURL,
		AccessTTL:               15 * time.Minute,
		RefreshTTL:              30 * 24 * time.Hour,
		InvitationTTL:           7 * 24 * time.Hour,
		AutoMigrate:             autoMigrate,
		MediaDir:                envOr("MEDIA_DIR", "./var/media"),
	}, nil
}

func validateActivationBaseURL(value, environment string) (string, error) {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Opaque != "" {
		return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL must be an absolute activation route")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery ||
		parsed.Fragment != "" || strings.ContainsAny(value, "?#") {
		return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL must not contain userinfo, query, or fragment")
	}
	switch parsed.Scheme {
	case "https":
		if parsed.Host == "" || parsed.Path != "/activate" || parsed.EscapedPath() != "/activate" {
			return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL must be https://<configured-origin>/activate")
		}
		if environment == "production" && parsed.Port() != "" {
			return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL must not use an explicit port in production")
		}
	case "belcanto":
		if environment == "production" {
			return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL must use https:// in production")
		}
		if parsed.Host != "activate" || parsed.Path != "" || parsed.Port() != "" || value != "belcanto://activate" {
			return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL custom scheme must be exactly belcanto://activate")
		}
	default:
		return "", fmt.Errorf("PUBLIC_ACTIVATION_BASE_URL must use https:// or belcanto://")
	}
	return parsed.String(), nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
