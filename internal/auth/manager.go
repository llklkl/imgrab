package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	configFileName          = "config.json"
	authMethodKeyring       = "keyring"
	authMethodEncryptedFile = "encrypted_file"
)

type authMethod struct {
	AuthMethod string `json:"auth_method"`
}

type Manager struct {
	mu      sync.RWMutex
	dir     string
	config  map[string]authMethod
	keyring *KeyringStorage
	file    *EncryptedFileStorage
}

func NewManager(dir string) *Manager {
	return &Manager{
		dir:     dir,
		config:  make(map[string]authMethod),
		keyring: NewKeyringStorage(),
		file:    NewEncryptedFileStorage(dir),
	}
}

func (m *Manager) Save(registry, username, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try keyring first
	err := m.keyring.Store(registry, username, password)
	if err == nil {
		m.config[registry] = authMethod{AuthMethod: authMethodKeyring}
		_ = m.saveConfig()
		return nil
	}

	// Fallback to encrypted file
	err = m.file.Store(registry, username, password)
	if err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	m.config[registry] = authMethod{AuthMethod: authMethodEncryptedFile}
	_ = m.saveConfig()
	return nil
}

func (m *Manager) Get(registry string) (string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_ = m.loadConfig()

	method, ok := m.config[registry]
	if !ok {
		// Try both storages
		u, p, err := m.keyring.Retrieve(registry)
		if err == nil {
			return u, p, nil
		}
		return m.file.Retrieve(registry)
	}

	switch method.AuthMethod {
	case authMethodKeyring:
		return m.keyring.Retrieve(registry)
	case authMethodEncryptedFile:
		return m.file.Retrieve(registry)
	default:
		return "", "", fmt.Errorf("unknown auth method: %s", method.AuthMethod)
	}
}

func (m *Manager) Delete(registry string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.loadConfig()

	method, ok := m.config[registry]
	if !ok {
		return nil
	}

	var err error
	switch method.AuthMethod {
	case authMethodKeyring:
		err = m.keyring.Remove(registry)
	case authMethodEncryptedFile:
		err = m.file.Remove(registry)
	}

	if err != nil {
		return err
	}

	delete(m.config, registry)
	return m.saveConfig()
}

func (m *Manager) List() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.loadConfig(); err != nil {
		return nil, err
	}

	registries := make([]string, 0, len(m.config))
	for r := range m.config {
		registries = append(registries, r)
	}
	return registries, nil
}

func (m *Manager) loadConfig() error {
	path := filepath.Join(m.dir, configFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var cfg struct {
		Registries map[string]authMethod `json:"registries"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	if cfg.Registries != nil {
		m.config = cfg.Registries
	}
	return nil
}

func (m *Manager) saveConfig() error {
	path := filepath.Join(m.dir, configFileName)

	cfg := struct {
		Registries map[string]authMethod `json:"registries"`
	}{
		Registries: m.config,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(m.dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GetAuthMethod returns the storage method for a registry
func (m *Manager) GetAuthMethod(registry string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_ = m.loadConfig()

	method, ok := m.config[registry]
	if !ok {
		return ""
	}
	return method.AuthMethod
}
