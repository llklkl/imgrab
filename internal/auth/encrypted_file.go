package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type EncryptedFileStorage struct {
	mu       sync.RWMutex
	dir      string
	key      []byte
	filename string
}

func NewEncryptedFileStorage(dir string) *EncryptedFileStorage {
	return &EncryptedFileStorage{
		dir:      dir,
		filename: "credentials.enc",
	}
}

func (e *EncryptedFileStorage) init() error {
	if e.key != nil {
		return nil
	}

	key, err := deriveKey()
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}
	e.key = key
	return nil
}

func (e *EncryptedFileStorage) Store(registry, username, password string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.init(); err != nil {
		return err
	}

	creds, err := e.loadAll()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if creds == nil {
		creds = make(map[string]credential)
	}

	creds[registry] = credential{Username: username, Password: password}

	return e.saveAll(creds)
}

func (e *EncryptedFileStorage) Retrieve(registry string) (string, string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if err := e.init(); err != nil {
		return "", "", err
	}

	creds, err := e.loadAll()
	if err != nil {
		return "", "", err
	}

	cred, ok := creds[registry]
	if !ok {
		return "", "", fmt.Errorf("credentials not found for %s", registry)
	}

	return cred.Username, cred.Password, nil
}

func (e *EncryptedFileStorage) Remove(registry string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.init(); err != nil {
		return err
	}

	creds, err := e.loadAll()
	if err != nil {
		return err
	}

	delete(creds, registry)

	return e.saveAll(creds)
}

func (e *EncryptedFileStorage) Exists(registry string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if err := e.init(); err != nil {
		return false
	}

	creds, err := e.loadAll()
	if err != nil {
		return false
	}

	_, ok := creds[registry]
	return ok
}

func (e *EncryptedFileStorage) loadAll() (map[string]credential, error) {
	path := filepath.Join(e.dir, e.filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	decrypted, err := Decrypt(data, e.key)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials: %w", err)
	}

	var creds map[string]credential
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, fmt.Errorf("unmarshal credentials: %w", err)
	}

	return creds, nil
}

func (e *EncryptedFileStorage) saveAll(creds map[string]credential) error {
	if err := os.MkdirAll(e.dir, 0700); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	encrypted, err := Encrypt(data, e.key)
	if err != nil {
		return fmt.Errorf("encrypt credentials: %w", err)
	}

	path := filepath.Join(e.dir, e.filename)
	return os.WriteFile(path, encrypted, 0600)
}
