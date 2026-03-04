package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/llklkl/imgrab/internal/tui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Docker images",
	Long:  `Search for Docker images on Docker Hub. If no query is provided, starts the TUI. If a query is provided, starts the TUI and searches for that query.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var initialQuery string
		if len(args) > 0 {
			initialQuery = args[0]
		}
		p := tea.NewProgram(tui.NewModel(initialQuery))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
