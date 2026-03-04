package cmd

import (
	"fmt"

	"github.com/llklkl/imgrab/internal/docker"
	"github.com/llklkl/imgrab/internal/registry"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:          "pull [image]",
	Short:        "Pull a Docker image",
	Long:         `Pull a Docker image from a registry and save it as a tar file.`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]
		outputDir, _ := cmd.Flags().GetString("output")
		arch, _ := cmd.Flags().GetString("arch")
		shouldImport, _ := cmd.Flags().GetBool("import")

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

		opts := &registry.PullOptions{
			OutputDir:    outputDir,
			ShowProgress: true,
		}

		outputPath, err := registry.SaveImageToTar(img, ref, opts)
		if err != nil {
			return fmt.Errorf("save image: %w", err)
		}

		fmt.Printf("\nImage saved to: %s\n", outputPath)

		if shouldImport {
			fmt.Println("\nImporting to Docker...")
			if err := docker.ImportTarToDocker(outputPath); err != nil {
				return fmt.Errorf("import to docker: %w", err)
			}
			fmt.Println("Successfully imported to Docker!")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringP("output", "o", "", "Output directory for the tar file")
	pullCmd.Flags().StringP("arch", "a", "", "Architecture to pull (default: current arch)")
	pullCmd.Flags().BoolP("import", "i", false, "Import the image to Docker after pulling")
}
