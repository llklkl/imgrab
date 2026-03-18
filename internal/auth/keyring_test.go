package auth

import (
	"os"
	"testing"
)

func TestKeyringStore(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping keyring test in CI")
	}

	k := NewKeyringStorage()
	registry := "test-registry.example.com"
	username := "testuser"
	password := "testpass"

	err := k.Store(registry, username, password)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}
	defer k.Remove(registry)

	u, p, err := k.Retrieve(registry)
	if err != nil {
		t.Fatalf("Retrieve() failed: %v", err)
	}

	if u != username || p != password {
		t.Errorf("expected (%s, %s), got (%s, %s)", username, password, u, p)
	}
}
