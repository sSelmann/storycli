// cmd/logs.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View logs for Story and Story-Geth services",
	Long:  `Provides subcommands to view real-time logs for Story and Story-Geth services.`,
}

var storyLogsCmd = &cobra.Command{
	Use:   "story",
	Short: "View logs for the Story service",
	RunE:  runServiceLogs("story"),
}

var gethLogsCmd = &cobra.Command{
	Use:   "geth",
	Short: "View logs for the Story-Geth service",
	RunE:  runServiceLogs("story-geth"),
}

var logsLines int

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.AddCommand(storyLogsCmd)
	logsCmd.AddCommand(gethLogsCmd)

	// Flags tanımlaması
	storyLogsCmd.Flags().IntVarP(&logsLines, "lines", "n", 20, "Number of log lines to display")
	gethLogsCmd.Flags().IntVarP(&logsLines, "lines", "n", 20, "Number of log lines to display")
}

func runServiceLogs(serviceName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		printInfo(fmt.Sprintf("Checking if '%s' service exists...", serviceName))

		exists, err := checkServiceExists(serviceName)
		if err != nil {
			return fmt.Errorf("failed to check if service exists: %w", err)
		}

		if !exists {
			printWarning(fmt.Sprintf("'%s' service is not found.", serviceName))
			return nil
		}

		printInfo(fmt.Sprintf("Fetching logs for '%s' service...", serviceName))
		if err := displayServiceLogs(serviceName, logsLines); err != nil {
			return fmt.Errorf("failed to display logs: %w", err)
		}

		return nil
	}
}

// displayServiceLogs displays the logs of a given systemd service
func displayServiceLogs(serviceName string, lines int) error {
	cmd := exec.Command("journalctl", "-fu", serviceName, "-o", "cat", "-n", strconv.Itoa(lines))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		printWarning(fmt.Sprintf("Failed to fetch logs for '%s' service.", serviceName))
		return err
	}

	return nil
}
