package auth

import (
	"testing"
)

func TestManagerSaveAndGet(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	registry := "index.docker.io"
	username := "testuser"
	password := "testpass"

	err := mgr.Save(registry, username, password)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	defer mgr.Delete(registry)

	u, p, err := mgr.Get(registry)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if u != username || p != password {
		t.Errorf("expected (%s, %s), got (%s, %s)", username, password, u, p)
	}
}

func TestManagerDelete(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	registry := "test.registry.com"

	_ = mgr.Save(registry, "user", "pass")
	err := mgr.Delete(registry)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	_, _, err = mgr.Get(registry)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
