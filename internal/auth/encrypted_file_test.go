package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedFileStorage(t *testing.T) {
	dir := t.TempDir()
	storage := NewEncryptedFileStorage(dir)

	registry := "test-registry"
	username := "testuser"
	password := "testpass"

	err := storage.Store(registry, username, password)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	u, p, err := storage.Retrieve(registry)
	if err != nil {
		t.Fatalf("Retrieve() failed: %v", err)
	}

	if u != username || p != password {
		t.Errorf("expected (%s, %s), got (%s, %s)", username, password, u, p)
	}

	err = storage.Remove(registry)
	if err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	if storage.Exists(registry) {
		t.Error("Exists() should return false after Remove()")
	}
}

func TestEncryptedFileStoragePersistence(t *testing.T) {
	dir := t.TempDir()
	storage1 := NewEncryptedFileStorage(dir)

	registry := "persist-registry"
	username := "persistuser"
	password := "persistpass"

	err := storage1.Store(registry, username, password)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Create new instance to test persistence
	storage2 := NewEncryptedFileStorage(dir)

	u, p, err := storage2.Retrieve(registry)
	if err != nil {
		t.Fatalf("Retrieve() from new instance failed: %v", err)
	}

	if u != username || p != password {
		t.Errorf("expected (%s, %s), got (%s, %s)", username, password, u, p)
	}

	// Clean up
	_ = storage2.Remove(registry)
}

func TestEncryptedFileStorageFileFormat(t *testing.T) {
	dir := t.TempDir()
	storage := NewEncryptedFileStorage(dir)

	registry := "format-registry"
	username := "formatuser"
	password := "formatpass"

	err := storage.Store(registry, username, password)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Check file exists
	filePath := filepath.Join(dir, "credentials.enc")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("credentials.enc file should exist")
	}

	// Check file is not plaintext
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("file should not be empty")
	}

	// Check version byte
	if data[0] != 1 {
		t.Errorf("expected version byte 1, got %d", data[0])
	}

	// Clean up
	_ = storage.Remove(registry)
}
