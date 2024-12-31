// cmd/restart.go
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Story and Story-Geth services",
	RunE:  runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
	pterm.Info.Printf("Restarting services...")

	services := []string{"story", "story-geth"}
	for _, service := range services {
		if err := performServiceAction(service, restartService); err != nil {
			return err
		}
	}

	pterm.Success.Printf("Services successfully restarted.")
	return nil
}

func restartService(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		printError(fmt.Sprintf("Failed to restart '%s' service: %v\nOutput: %s", serviceName, err, string(output)))
		return err
	}
	return nil
}

func performServiceAction(serviceName string, action func(string) error) error {
	exists, err := checkServiceExists(serviceName)
	if err != nil {
		return fmt.Errorf("failed to check if service '%s' exists: %w", serviceName, err)
	}
	if !exists {
		pterm.Warning.Printf(fmt.Sprintf("'%s' service is not installed.", serviceName))
		return nil
	}
	pterm.Info.Printf(fmt.Sprintf("Performing action on '%s' service...", serviceName))
	if err := action(serviceName); err != nil {
		return err
	}
	return nil
}
