package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "imgrab",
	Short: "imgrab - Docker image pull CLI tool",
	Long:  `imgrab is a CLI tool for pulling Docker images from Docker Hub and private registries.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
