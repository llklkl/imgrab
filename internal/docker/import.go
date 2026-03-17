package docker

import (
	"fmt"
	"os"
	"os/exec"
)

func ImportTarToDocker(tarPath string) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker command not found: %w", err)
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("open tar file: %w", err)
	}
	defer file.Close()

	cmd := exec.Command("docker", "load", "-i", tarPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker load failed: %w", err)
	}

	return nil
}
