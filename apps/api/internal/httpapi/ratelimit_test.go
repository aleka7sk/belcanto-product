package httpapi

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterHasHardCardinalityCap(t *testing.T) {
	limiter := newRateLimiter()
	now := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	for index := 0; index < maxRateLimitBuckets; index++ {
		if !limiter.allow(fmt.Sprintf("attacker-token-%d", index), 1, time.Minute) {
			t.Fatalf("bucket %d unexpectedly denied before cap", index)
		}
	}
	if limiter.allow("attacker-token-over-cap", 1, time.Minute) {
		t.Fatal("new attacker-controlled key was accepted over hard cap")
	}
	if len(limiter.buckets) != maxRateLimitBuckets {
		t.Fatalf("bucket count = %d, want %d", len(limiter.buckets), maxRateLimitBuckets)
	}
	if limiter.allow("attacker-token-0", 1, time.Minute) {
		t.Fatal("existing exhausted bucket unexpectedly refilled")
	}

	now = now.Add(2 * time.Hour)
	if !limiter.allow("legitimate-new-key", 1, time.Minute) {
		t.Fatal("stale buckets were not pruned after inactivity")
	}
	if len(limiter.buckets) > maxRateLimitBuckets {
		t.Fatalf("bucket count exceeded hard cap: %d", len(limiter.buckets))
	}
}

func TestSensitiveLimiterHasCoarseIPLayerAcrossRotatingSubjects(t *testing.T) {
	api := &API{limits: newRateLimiter()}
	request := httptest.NewRequest("POST", "https://api.test/v1/sessions", nil)
	request.RemoteAddr = "203.0.113.10:12345"
	for index := 0; index < 3; index++ {
		if !api.allowSensitive(request, "test_sign_in", fmt.Sprintf("rotating-subject-%d", index), 10, 3) {
			t.Fatalf("request %d denied before coarse IP capacity", index)
		}
	}
	if api.allowSensitive(request, "test_sign_in", "fresh-rotating-subject", 10, 3) {
		t.Fatal("rotating attacker subject bypassed coarse IP capacity")
	}
}

func TestSensitiveLimiterSubjectCannotEscapeByRotatingIP(t *testing.T) {
	api := &API{limits: newRateLimiter()}
	request := httptest.NewRequest("POST", "https://api.test/v1/sessions", nil)
	for index := 0; index < 3; index++ {
		request.RemoteAddr = fmt.Sprintf("203.0.113.%d:12345", index+1)
		if !api.allowSensitive(request, "test_subject", "same-account", 3, 10) {
			t.Fatalf("request %d denied before subject capacity", index)
		}
	}
	request.RemoteAddr = "198.51.100.99:12345"
	if api.allowSensitive(request, "test_subject", "same-account", 3, 10) {
		t.Fatal("rotating IP bypassed subject capacity")
	}
}
