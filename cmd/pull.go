package cmd

import (
	"fmt"
	"os"

	"github.com/llklkl/imgrab/internal/docker"
	"github.com/llklkl/imgrab/internal/registry"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [image]",
	Short: "Pull a Docker image",
	Long: `Pull a Docker image from a registry and import it to Docker by default.

Examples:
  # Pull and import nginx latest (default behavior)
  imgrab pull nginx

  # Pull specific version and import
  imgrab pull nginx:1.25.3

  # Pull only, don't import to Docker
  imgrab pull nginx --download-only

  # Pull only and save to specific directory
  imgrab pull nginx --download-only -o ./images

  # Specify architecture
  imgrab pull nginx -a arm64`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]
		outputDir, _ := cmd.Flags().GetString("output")
		arch, _ := cmd.Flags().GetString("arch")
		downloadOnly, _ := cmd.Flags().GetBool("download-only")

		fmt.Printf("Pulling image: %s\n", imageRef)

		ref, err := registry.ParseImageRef(imageRef, arch, "")
		if err != nil {
			return fmt.Errorf("parse image reference: %w", err)
		}

		auth, err := registry.GetCredential(ref.Registry)
		if err != nil {
			return fmt.Errorf("get credential: %w", err)
		}

		client := registry.NewClient().WithAuth(auth)
		img, err := client.PullImage(ref)
		if err != nil {
			return fmt.Errorf("pull image: %w", err)
		}

		// Use temp directory if importing, otherwise use specified output
		targetDir := outputDir
		if !downloadOnly {
			tempDir, err := os.MkdirTemp("", "imgrab-*")
			if err != nil {
				return fmt.Errorf("create temp directory: %w", err)
			}
			targetDir = tempDir
		}

		opts := &registry.PullOptions{
			OutputDir:    targetDir,
			ShowProgress: true,
		}

		outputPath, err := registry.SaveImageToTar(img, ref, opts)
		if err != nil {
			return fmt.Errorf("save image: %w", err)
		}

		// Clean up temp file after import
		cleanup := func() {
			if !downloadOnly {
				os.Remove(outputPath)
				os.Remove(targetDir)
			}
		}

		if downloadOnly {
			fmt.Printf("\nImage saved to: %s\n", outputPath)
			return nil
		}

		fmt.Println("\nImporting to Docker...")
		if err := docker.ImportTarToDocker(outputPath); err != nil {
			cleanup()
			return fmt.Errorf("import to docker: %w", err)
		}
		fmt.Println("Successfully imported to Docker!")

		cleanup()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringP("output", "o", "", "Output directory for the tar file (only with --download-only)")
	pullCmd.Flags().StringP("arch", "a", "", "Architecture to pull (default: current arch)")
	pullCmd.Flags().BoolP("download-only", "d", false, "Download only, don't import to Docker")
}
