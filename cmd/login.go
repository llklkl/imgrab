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
	Long:  `Login to a Docker registry and save credentials.`,
	Args:  cobra.MaximumNArgs(1),
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

		if err := registry.SaveCredential(registryAddr, username, password); err != nil {
			return fmt.Errorf("save credential: %w", err)
		}

		fmt.Printf("Login successful for %s\n", registryAddr)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "Username for login")
	loginCmd.Flags().StringP("password", "p", "", "Password for login")
}
