package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

// ValidateCredentials validates credentials against a registry
func ValidateCredentials(registry, username, password string) error {
	auth := authn.FromConfig(authn.AuthConfig{
		Username: username,
		Password: password,
	})

	// Try to list catalog to verify credentials
	// For Docker Hub, this will work; for private registries, it depends on configuration
	reg, err := name.NewRegistry(registry)
	if err != nil {
		return fmt.Errorf("invalid registry: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to get catalog (this requires authentication on most registries)
	_, err = remote.Catalog(ctx, reg, remote.WithAuth(auth))
	if err != nil {
		// Catalog might fail due to permissions, but we can still check the error type
		// If it's an auth error, return that; otherwise credentials might be valid
		// but catalog endpoint is not accessible
		if isAuthError(err) {
			return fmt.Errorf("authentication failed: invalid credentials")
		}
		// For other errors (e.g., catalog not supported), we assume credentials are valid
		// as the error is not an auth error
	}

	return nil
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for common auth error patterns
	return strings.Contains(errStr, "UNAUTHORIZED") ||
		strings.Contains(errStr, "authentication required") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "denied")
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
