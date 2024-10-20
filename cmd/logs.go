package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View logs for Story and Story-Geth services",
	Long:  `Provides subcommands to view real-time logs for Story and Story-Geth services.`,
}

// storyLogsCmd represents the 'logs story' subcommand
var storyLogsCmd = &cobra.Command{
	Use:   "story",
	Short: "View logs for the Story service",
	RunE:  runStoryLogs,
}

// gethLogsCmd represents the 'logs geth' subcommand
var gethLogsCmd = &cobra.Command{
	Use:   "geth",
	Short: "View logs for the Story-Geth service",
	RunE:  runGethLogs,
}

var (
	logsLines int
)

func init() {
	// Add 'logs' command to root
	rootCmd.AddCommand(logsCmd)

	// Add subcommands to 'logs'
	logsCmd.AddCommand(storyLogsCmd)
	logsCmd.AddCommand(gethLogsCmd)

	// Define flags for subcommands
	storyLogsCmd.Flags().IntVarP(&logsLines, "lines", "n", 20, "Number of log lines to display")
	gethLogsCmd.Flags().IntVarP(&logsLines, "lines", "n", 20, "Number of log lines to display")
}

// runStoryLogs executes the 'logs story' subcommand
func runStoryLogs(cmd *cobra.Command, args []string) error {
	serviceName := "story"
	printInfo(fmt.Sprintf("Checking if '%s' service exists...", serviceName))

	exists, err := checkServiceExists(serviceName)
	if err != nil {
		return err
	}

	if !exists {
		printWarning(fmt.Sprintf("'%s' service is not found.", serviceName))
		return nil
	}

	printInfo(fmt.Sprintf("Fetching logs for '%s' service...", serviceName))
	err = displayServiceLogs(serviceName, logsLines)
	if err != nil {
		return err
	}

	return nil
}

// runGethLogs executes the 'logs geth' subcommand
func runGethLogs(cmd *cobra.Command, args []string) error {
	serviceName := "story-geth"
	printInfo(fmt.Sprintf("Checking if '%s' service exists...", serviceName))

	exists, err := checkServiceExists(serviceName)
	if err != nil {
		return err
	}

	if !exists {
		printWarning(fmt.Sprintf("'%s' service is not installed.", serviceName))
		return nil
	}

	printInfo(fmt.Sprintf("Fetching logs for '%s' service...", serviceName))
	err = displayServiceLogs(serviceName, logsLines)
	if err != nil {
		return err
	}

	return nil
}

// displayServiceLogs displays the logs of a given systemd service
func displayServiceLogs(serviceName string, lines int) error {
	linesStr := strconv.Itoa(lines)
	cmdStr := fmt.Sprintf("journalctl -fu %s -o cat -n %s", serviceName, linesStr)
	cmd := exec.Command("bash", "-c", cmdStr)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		printWarning(fmt.Sprintf("Failed to fetch logs for '%s' service.", serviceName))
		return err
	}

	return nil
}
