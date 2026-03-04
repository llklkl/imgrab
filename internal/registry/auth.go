package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthConfig struct {
	Credentials map[string]Credential `json:"credentials"`
}

const (
	configDirName  = ".imgrab"
	configFileName = "config.json"
)

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	configDir := filepath.Join(homeDir, configDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return filepath.Join(configDir, configFileName), nil
}

func loadAuthConfig() (*AuthConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthConfig{
				Credentials: make(map[string]Credential),
			}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if config.Credentials == nil {
		config.Credentials = make(map[string]Credential)
	}

	return &config, nil
}

func saveAuthConfig(config *AuthConfig) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func SaveCredential(registry, username, password string) error {
	config, err := loadAuthConfig()
	if err != nil {
		return err
	}

	config.Credentials[registry] = Credential{
		Username: username,
		Password: password,
	}

	return saveAuthConfig(config)
}

func GetCredential(registry string) (authn.Authenticator, error) {
	config, err := loadAuthConfig()
	if err != nil {
		return nil, err
	}

	cred, ok := config.Credentials[registry]
	if !ok {
		return authn.Anonymous, nil
	}

	return authn.FromConfig(authn.AuthConfig{
		Username: cred.Username,
		Password: cred.Password,
	}), nil
}


