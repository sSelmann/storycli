// cmd/stop.go
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Story and Story-Geth services",
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	pterm.Info.Printf("Stopping services...")

	services := []string{"story", "story-geth"}
	for _, service := range services {
		if err := performServiceAction(service, stopService); err != nil {
			return err
		}
	}

	pterm.Success.Printf("Services successfully stopped.")
	return nil
}

func stopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		printError(fmt.Sprintf("Failed to stop '%s' service: %v\nOutput: %s", serviceName, err, string(output)))
		return err
	}
	return nil
}
