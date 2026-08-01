package security

import (
	"bytes"
	"testing"
)

func TestInvitationTokenIsDeterministicOnlyForSameMasterAndIdentifier(t *testing.T) {
	first, err := NewTokenCodec(bytes.Repeat([]byte{0x11}, 32))
	if err != nil {
		t.Fatalf("new first codec: %v", err)
	}
	second, err := NewTokenCodec(bytes.Repeat([]byte{0x22}, 32))
	if err != nil {
		t.Fatalf("new second codec: %v", err)
	}
	token := first.InvitationToken("inv_0123456789abcdef0123456789abcdef")
	if len(token) != 43 {
		t.Fatalf("token length = %d, want 43", len(token))
	}
	if token != first.InvitationToken("inv_0123456789abcdef0123456789abcdef") {
		t.Fatal("same key and invitation ID must reproduce the token")
	}
	if token == first.InvitationToken("inv_fedcba9876543210fedcba9876543210") {
		t.Fatal("different invitation ID reproduced the token")
	}
	if token == second.InvitationToken("inv_0123456789abcdef0123456789abcdef") {
		t.Fatal("different master key reproduced the token")
	}
	if bytes.Contains(first.Digest(token), []byte(token)) {
		t.Fatal("stored digest contains the raw token")
	}
}

func TestNewTokenCodecRequiresStrongMasterKey(t *testing.T) {
	if _, err := NewTokenCodec(make([]byte, 31)); err == nil {
		t.Fatal("31-byte master key must be rejected")
	}
}
