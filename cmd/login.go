package cmd

import (
	"fmt"
	"syscall"

	"github.com/llklkl/imgrab/internal/registry"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login [registry]",
	Short: "Login to a Docker registry",
	Long: `Login to a Docker registry and save credentials securely.

If no registry is provided, defaults to Docker Hub (index.docker.io).
If username or password are not provided via flags, they will be prompted interactively.

Examples:
  # Login to Docker Hub
  imgrab login

  # Login to Docker Hub with credentials
  imgrab login -u your_username -p your_password

  # Login to private registry
  imgrab login registry.example.com

  # Login to private registry with credentials
  imgrab login registry.example.com -u your_username -p your_password`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryAddr := "index.docker.io"
		if len(args) > 0 {
			registryAddr = args[0]
		}

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if username == "" {
			fmt.Print("Username: ")
			if _, err := fmt.Scanln(&username); err != nil {
				return fmt.Errorf("read username: %w", err)
			}
		}

		if password == "" {
			fmt.Print("Password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			password = string(bytePassword)
			fmt.Println()
		}

		// Validate credentials before saving
		fmt.Println("Validating credentials...")
		if err := registry.ValidateCredentials(registryAddr, username, password); err != nil {
			return fmt.Errorf("credential validation failed: %w", err)
		}

		if err := registry.SaveCredential(registryAddr, username, password); err != nil {
			return fmt.Errorf("save credential: %w", err)
		}

		// Show storage method
		authMethod := registry.GetAuthMethod(registryAddr)
		storageMethod := "encrypted file"
		if authMethod == "keyring" {
			storageMethod = "system keyring"
		}

		fmt.Printf("Login successful for %s\n", registryAddr)
		fmt.Printf("Credentials stored using %s\n", storageMethod)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "Username for login")
	loginCmd.Flags().StringP("password", "p", "", "Password for login")
}
