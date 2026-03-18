package registry

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/llklkl/imgrab/internal/auth"
)

var (
	managerOnce sync.Once
	authManager *auth.Manager
)

func getAuthManager() *auth.Manager {
	managerOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".config", "imgrab")
		authManager = auth.NewManager(configDir)
	})
	return authManager
}

func SaveCredential(registry, username, password string) error {
	return getAuthManager().Save(registry, username, password)
}

func GetCredential(registry string) (authn.Authenticator, error) {
	mgr := getAuthManager()
	username, password, err := mgr.Get(registry)
	if err != nil {
		return authn.Anonymous, nil
	}

	return authn.FromConfig(authn.AuthConfig{
		Username: username,
		Password: password,
	}), nil
}

func DeleteCredential(registry string) error {
	return getAuthManager().Delete(registry)
}

func ListCredentials() ([]string, error) {
	return getAuthManager().List()
}

func GetAuthMethod(registry string) string {
	return getAuthManager().GetAuthMethod(registry)
}
