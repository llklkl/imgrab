package cmd

import "github.com/spf13/cobra"

var loginCmd = &cobra.Command{
	Use:   "login [registry]",
	Short: "Login to a Docker registry",
	Long:  `Login to a Docker registry and save credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "Username for login")
	loginCmd.Flags().StringP("password", "p", "", "Password for login")
}
