// cmd/status.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of Story and Story-Geth services",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	printInfo("Checking service statuses...")

	services := []string{"story", "story-geth"}
	for _, service := range services {
		if err := performServiceAction(service, displayServiceStatus); err != nil {
			return err
		}
	}

	return nil
}

func displayServiceStatus(serviceName string) error {
	cmd := exec.Command("systemctl", "status", serviceName, "--no-pager", "-n", "3")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		printWarning(fmt.Sprintf("Failed to get status for '%s' service.", serviceName))
		return err
	}

	return nil
}
