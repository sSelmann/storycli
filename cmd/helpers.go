package cmd

import (
	"bytes"
	"os/exec"
)

// checkServiceExists checks if a given systemd service exists
func checkServiceExists(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return false, nil
	}

	return bytes.Contains(out.Bytes(), []byte(serviceName+".service")), nil
}
