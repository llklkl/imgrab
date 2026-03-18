package auth

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const service = "imgrab"

type KeyringStorage struct{}

func NewKeyringStorage() *KeyringStorage {
	return &KeyringStorage{}
}

func (k *KeyringStorage) Store(registry, username, password string) error {
	cred := credential{Username: username, Password: password}
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("marshal credential: %w", err)
	}
	return keyring.Set(service, registry, string(data))
}

func (k *KeyringStorage) Retrieve(registry string) (string, string, error) {
	data, err := keyring.Get(service, registry)
	if err != nil {
		return "", "", err
	}

	var cred credential
	if err := json.Unmarshal([]byte(data), &cred); err != nil {
		return "", "", fmt.Errorf("unmarshal credential: %w", err)
	}

	return cred.Username, cred.Password, nil
}

func (k *KeyringStorage) Remove(registry string) error {
	return keyring.Delete(service, registry)
}

func (k *KeyringStorage) Exists(registry string) bool {
	_, err := keyring.Get(service, registry)
	return err == nil
}
