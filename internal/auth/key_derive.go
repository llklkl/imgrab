package auth

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
)

func deriveKey() ([]byte, error) {
	var machineID string
	var err error

	switch runtime.GOOS {
	case "linux":
		machineID, err = getLinuxMachineID()
	case "darwin":
		machineID, err = getDarwinMachineID()
	case "windows":
		machineID, err = getWindowsMachineID()
	default:
		machineID = getFallbackMachineID()
	}

	if err != nil {
		machineID = getFallbackMachineID()
	}

	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	homeDir, _ := os.UserHomeDir()

	combined := fmt.Sprintf("%s:%s:%s", machineID, username, homeDir)
	hash := sha256.Sum256([]byte(combined))

	key := make([]byte, 32)
	copy(key, hash[:])
	return key, nil
}

func getLinuxMachineID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		data, err = os.ReadFile("/var/lib/dbus/machine-id")
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(data)), nil
}

func getDarwinMachineID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ioreg: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				uuid := strings.TrimSpace(strings.Trim(parts[1], "\""))
				return uuid, nil
			}
		}
	}

	return "", fmt.Errorf("IOPlatformUUID not found")
}

func getWindowsMachineID() (string, error) {
	cmd := exec.Command("reg", "query", "HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Cryptography", "/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("reg query: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "MachineGuid") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[len(fields)-1], nil
			}
		}
	}

	return "", fmt.Errorf("MachineGuid not found")
}

func getFallbackMachineID() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "fallback-default-id"
	}
	hash := sha256.Sum256([]byte(info.String()))
	return fmt.Sprintf("%x", hash)
}
