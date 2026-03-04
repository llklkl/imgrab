package cmd

import "github.com/spf13/cobra"

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Docker images",
	Long:  `Search for Docker images on Docker Hub with a TUI interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
