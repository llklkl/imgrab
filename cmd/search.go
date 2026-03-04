package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/llklkl/imgrab/internal/registry"
	"github.com/llklkl/imgrab/internal/tui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Docker images",
	Long:  `Search for Docker images on Docker Hub. If no query is provided, starts the TUI.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			p := tea.NewProgram(tui.NewModel())
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

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
			if item.RepoOwner != "" {
				fmt.Printf("   Owner: %s\n", item.RepoOwner)
			}
			if item.Description != "" {
				fmt.Printf("   %s\n", item.Description)
			}
			badge := ""
			if item.IsOfficial {
				badge += " [OFFICIAL]"
			}
			if item.IsAutomated {
				badge += " [AUTOMATED]"
			}
			fmt.Printf("   Stars: %d | Pulls: %s%s\n\n", item.Stars, registry.FormatNumber(item.PullCount), badge)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
