package cmd

import (
	"fmt"

	"github.com/llklkl/imgrab/internal/registry"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout [registry]",
	Short: "Logout from a Docker registry",
	Long: `Logout from a Docker registry by removing stored credentials.

If no registry is provided, defaults to Docker Hub (index.docker.io).

Examples:
  # Logout from Docker Hub
  imgrab logout

  # Logout from private registry
  imgrab logout registry.example.com`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryAddr := "index.docker.io"
		if len(args) > 0 {
			registryAddr = args[0]
		}

		if err := registry.DeleteCredential(registryAddr); err != nil {
			return fmt.Errorf("logout: %w", err)
		}

		fmt.Printf("Logged out from %s\n", registryAddr)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
