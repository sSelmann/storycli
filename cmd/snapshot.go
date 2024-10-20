package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Download and apply snapshots for Story node",
	Long: `Downloads and applies snapshots from selected providers (Itrocket or Krews)
to initialize or update your Story node.`,
	RunE: runSnapshot,
}

var (
	// homeDirFlag holds the value for the --home flag
	homeDirFlag string
)

func init() {
	// Add 'snapshot' command to root
	rootCmd.AddCommand(snapshotCmd)

	// Define the --home flag for snapshot command
	snapshotCmd.Flags().StringVar(&homeDirFlag, "home", defaultHomeDir(), "Home directory for Story node")
}

// defaultHomeDir returns the default home directory path
func defaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		os.Exit(1)
	}
	return filepath.Join(home)
}

// runSnapshot executes the snapshot command
func runSnapshot(cmd *cobra.Command, args []string) error {
	// Get the home directory from flag or default
	homeDir := homeDirFlag

	err := fetchSnapshotSizes()
	if err != nil {
		printWarning(fmt.Sprintf("Failed to fetch snapshot sizes: %v", err))
	}

	// Pruning mode info message
	printInfo("\nPruning Mode Information:")
	fmt.Printf(" - Pruned Mode: Stores only recent blockchain data, reducing disk usage. Snapshot size: %s\n", prunedSnapshotSize)
	fmt.Printf(" - Archive Mode: Stores the entire blockchain history, requiring more disk space. Snapshot size: %s\n\n", archiveSnapshotSize)

	if pruningMode == "" {
		prompt := promptui.Select{
			Label: "Select the pruning mode",
			Items: []string{"pruned", "archive"},
		}
		_, pruningMode, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	// Prompt the user to select snapshot provider
	provider, err := selectSnapshotProvider()
	if err != nil {
		return err
	}

	// Enable and start services
	printInfo("stopping services...")
	err = runCommand("sudo systemctl stop story story-geth")
	if err != nil {
		return err
	}

	// Execute the corresponding snapshot download function
	switch provider {
	case "Itrocket":
		err = downloadSnapshotItrocket(homeDir, pruningMode)
	case "Krews":
		err = downloadSnapshotKrews(homeDir, pruningMode)
	default:
		err = errors.New("provider not found")
	}

	if err != nil {
		return err
	}

	// Print restart services info
	pterm.Info.Printf(
		"run %s to restart the systemd services and apply the new values\n",
		pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprintf("scli restart"),
	)

	return nil
}

// selectSnapshotProvider prompts the user to choose between Itrocket and Krews using promptui
func selectSnapshotProvider() (string, error) {
	providerOptions := []string{"Itrocket", "Krews"}

	prompt := promptui.Select{
		Label: "Select the snapshot provider",
		Items: providerOptions,
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}
