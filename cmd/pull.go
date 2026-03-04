package cmd

import "github.com/spf13/cobra"

var pullCmd = &cobra.Command{
	Use:   "pull [image]",
	Short: "Pull a Docker image",
	Long:  `Pull a Docker image from a registry and save it as a tar file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringP("output", "o", "", "Output directory for the tar file")
	pullCmd.Flags().StringP("arch", "a", "", "Architecture to pull (default: current arch)")
	pullCmd.Flags().BoolP("import", "i", false, "Import the image to Docker after pulling")
}
