package cmd

import (
	"fmt"

	"github.com/llklkl/imgrab/internal/registry"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Docker images",
	Long:  `Search for Docker images on Docker Hub.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		fmt.Printf("Searching for: %s\n\n", query)

		result, err := registry.SearchImages(query, 1, 10)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(result.Results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("Found %d results:\n\n", result.Count)
		for i, item := range result.Results {
			fmt.Printf("%d. %s\n", i+1, item.Name)
			if item.Description != "" {
				fmt.Printf("   %s\n", item.Description)
			}
			fmt.Printf("   Stars: %d", item.Stars)
			if item.IsOfficial {
				fmt.Printf(" [OFFICIAL]")
			}
			if item.IsAutomated {
				fmt.Printf(" [AUTOMATED]")
			}
			fmt.Println()
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
