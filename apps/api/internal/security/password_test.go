package security

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPasswordHasherUnicodeBoundariesAndNormalization(t *testing.T) {
	hasher := NewPasswordHasher()
	if _, err := hasher.Hash(strings.Repeat("я", 14)); err == nil {
		t.Fatal("14 Unicode characters must be rejected")
	}
	minimum := strings.Repeat("я", 15)
	minimumHash, err := hasher.Hash(minimum)
	if err != nil {
		t.Fatalf("hash minimum password: %v", err)
	}
	if !hasher.Verify(minimum, minimumHash) {
		t.Fatal("minimum password did not verify")
	}
	maximum := strings.Repeat("界", 128)
	maximumHash, err := hasher.Hash(maximum)
	if err != nil {
		t.Fatalf("hash maximum password: %v", err)
	}
	if !hasher.Verify(maximum, maximumHash) {
		t.Fatal("maximum password did not verify")
	}
	if _, err := hasher.Hash(strings.Repeat("界", 129)); err == nil {
		t.Fatal("129 Unicode characters must be rejected")
	}

	decomposed := strings.Repeat("e\u0301", 15)
	composed := strings.Repeat("é", 15)
	hash, err := hasher.Hash(decomposed)
	if err != nil {
		t.Fatalf("hash normalized password: %v", err)
	}
	if !hasher.Verify(composed, hash) {
		t.Fatal("canonically equivalent NFC password did not verify")
	}
}

func TestArgon2WorkIsProcessWideAndBounded(t *testing.T) {
	if cap(argon2Slots) != argon2Concurrency || argon2Concurrency < 1 {
		t.Fatalf("Argon2 semaphore capacity = %d", cap(argon2Slots))
	}
	for index := 0; index < argon2Concurrency; index++ {
		argon2Slots <- struct{}{}
	}
	t.Cleanup(func() {
		for len(argon2Slots) > 0 {
			<-argon2Slots
		}
	})
	done := make(chan error, 1)
	go func() {
		_, err := NewPasswordHasher().Hash("Unique semaphore password 2026")
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("Argon2 work bypassed full process semaphore: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	<-argon2Slots
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("hash after semaphore release: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Argon2 work did not resume after semaphore release")
	}
}

func TestPasswordHasherRejectsCommonAndMalformedValues(t *testing.T) {
	hasher := NewPasswordHasher()
	if _, err := hasher.Hash("123456789012345"); err == nil {
		t.Fatal("common password must be rejected")
	}
	if _, err := hasher.Hash("correcthorsebatterystaple"); err == nil {
		t.Fatal("embedded common-password corpus entry must be rejected")
	}
	if hasher.Verify("any password long enough", "not-a-phc-string") {
		t.Fatal("malformed PHC string must not verify")
	}
	if _, err := hasher.VerifyCredential("any password long enough", "not-a-phc-string"); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("malformed stored PHC error = %v", err)
	}
	if hasher.Verify("unknown account password", hasher.DummyHash()) {
		t.Fatal("dummy hash unexpectedly verified")
	}
}
