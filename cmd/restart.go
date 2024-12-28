// cmd/restart.go
package cmd

import (
	"fmt"
	"os/exec"

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
	printInfo("Restarting services...")

	services := []string{"story", "story-geth"}
	for _, service := range services {
		if err := performServiceAction(service, restartService); err != nil {
			return err
		}
	}

	printSuccess("Services successfully restarted.")
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
		printWarning(fmt.Sprintf("'%s' service is not installed.", serviceName))
		return nil
	}
	printInfo(fmt.Sprintf("Performing action on '%s' service...", serviceName))
	if err := action(serviceName); err != nil {
		return err
	}
	return nil
}
