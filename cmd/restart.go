// cmd/restart.go
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/manifoldco/promptui"
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

	restartCmd.Flags().Bool("all", false, "Restart story and story-geth services together")
}

func runRestart(cmd *cobra.Command, args []string) error {
	pterm.Info.Println("Restarting services...\n")

	// Boolean değeri kontrol ediyoruz
	isAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		return fmt.Errorf("failed to parse 'all' flag: %w", err)
	}

	if isAll {
		restartBoth()
	} else {

		pmPrompt := promptui.Select{
			Label: "Select restarting service",
			Items: []string{"story", "story-geth", "both"},
		}
		_, pmResult, err := pmPrompt.Run()
		if err != nil {
			return err
		}
		var service = pmResult

		if service == "both" {
			restartBoth()
		} else {
			restartService(service)
		}
	}

	pterm.Success.Println("Services successfully restarted.")
	return nil
}

func restartService(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		printError(fmt.Sprintln("Failed to restart '%s' service: %v\nOutput: %s", serviceName, err, string(output)))
		return err
	}
	return nil
}

func restartBoth() error {
	if err := restartService("story"); err != nil {
		return err
	}
	if err := restartService("story-geth"); err != nil {
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
		pterm.Warning.Println(fmt.Sprintf("'%s' service is not installed.", serviceName))
		return nil
	}
	pterm.Info.Println(fmt.Sprintf("Performing action on '%s' service...", serviceName))
	if err := action(serviceName); err != nil {
		return err
	}
	return nil
}
