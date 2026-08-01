package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestLoadRequiresDatabaseAndSafeActivationBaseURL(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	t.Setenv("TOKEN_HMAC_KEY_BASE64", key)
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("missing DATABASE_URL must fail startup")
	}

	t.Setenv("DATABASE_URL", "postgres://localhost/belcanto")
	t.Setenv("PUBLIC_ACTIVATION_BASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("missing PUBLIC_ACTIVATION_BASE_URL must fail startup")
	}
	for _, invalid := range []string{
		"http://app.example.test/activate",
		"https://user@app.example.test/activate",
		"https://app.example.test/activate?source=secret",
		"https://app.example.test/activate#token=bad",
		"https://app.example.test/activate/",
		"https://app.example.test/account/activate",
		"https://app.example.test/%61ctivate",
		"belcanto://activate/",
		"belcanto://activate/account",
		"belcanto://other",
		"/relative/activate",
	} {
		t.Setenv("PUBLIC_ACTIVATION_BASE_URL", invalid)
		if _, err := Load(); err == nil {
			t.Fatalf("unsafe activation base URL %q was accepted", invalid)
		}
	}

	t.Setenv("PUBLIC_ACTIVATION_BASE_URL", "https://app.example.test/activate")
	configuration, err := Load()
	if err != nil {
		t.Fatalf("load valid https configuration: %v", err)
	}
	if configuration.PublicActivationBaseURL != "https://app.example.test/activate" {
		t.Fatalf("normalized activation URL = %q", configuration.PublicActivationBaseURL)
	}
	t.Setenv("PUBLIC_ACTIVATION_BASE_URL", "belcanto://activate")
	if _, err := Load(); err != nil {
		t.Fatalf("valid custom activation scheme rejected: %v", err)
	}
	t.Setenv("APP_ENV", "test")
	if _, err := Load(); err != nil {
		t.Fatalf("test environment rejected exact custom activation scheme: %v", err)
	}

	t.Setenv("APP_ENV", "production")
	if _, err := Load(); err == nil {
		t.Fatal("production accepted a custom-scheme activation URL")
	}
	t.Setenv("PUBLIC_ACTIVATION_BASE_URL", "https://app.example.test:8443/activate")
	if _, err := Load(); err == nil {
		t.Fatal("production accepted an activation URL with an explicit port")
	}
	t.Setenv("PUBLIC_ACTIVATION_BASE_URL", "https://app.example.test/activate")
	configuration, err = Load()
	if err != nil {
		t.Fatalf("production rejected exact HTTPS activation URL: %v", err)
	}
	if configuration.Environment != "production" {
		t.Fatalf("environment = %q, want production", configuration.Environment)
	}

	t.Setenv("APP_ENV", "staging")
	if _, err := Load(); err == nil {
		t.Fatal("unknown APP_ENV was accepted")
	}
}

func TestLoadDefaultsAPPEnvToDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_URL", "postgres://localhost/belcanto")
	t.Setenv("TOKEN_HMAC_KEY_BASE64", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))
	t.Setenv("PUBLIC_ACTIVATION_BASE_URL", "belcanto://activate")
	configuration, err := Load()
	if err != nil {
		t.Fatalf("load default development configuration: %v", err)
	}
	if configuration.Environment != "development" {
		t.Fatalf("environment = %q, want development", configuration.Environment)
	}
}
