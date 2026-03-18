package auth

import (
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key, err := deriveKey()
	if err != nil {
		t.Fatalf("deriveKey() failed: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("expected key length 32, got %d", len(key))
	}
}

func TestDeriveKeyConsistency(t *testing.T) {
	key1, err := deriveKey()
	if err != nil {
		t.Fatalf("first deriveKey() failed: %v", err)
	}

	key2, err := deriveKey()
	if err != nil {
		t.Fatalf("second deriveKey() failed: %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("deriveKey() should return consistent keys")
	}
}
