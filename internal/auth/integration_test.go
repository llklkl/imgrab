//go:build integration
// +build integration

package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationManager(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "imgrab_test")
	defer os.RemoveAll(configDir)

	mgr := NewManager(configDir)

	// Test save and retrieve
	err = mgr.Save("registry.example.com", "testuser", "testpass")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	username, password, err := mgr.Get("registry.example.com")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if username != "testuser" || password != "testpass" {
		t.Errorf("expected (testuser, testpass), got (%s, %s)", username, password)
	}

	// Test list
	registries, err := mgr.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(registries) != 1 || registries[0] != "registry.example.com" {
		t.Errorf("expected [registry.example.com], got %v", registries)
	}

	// Test delete
	err = mgr.Delete("registry.example.com")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, _, err = mgr.Get("registry.example.com")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
