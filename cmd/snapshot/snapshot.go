package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/utils/config"
	"github.com/spf13/cobra"
)

var (
	// homeDirFlag is used to store the home directory path (passed via --home).
	homeDirFlag string

	// pruningMode represents the selected pruning mode ("pruned" or "archive").
	pruningMode string

	// selectedProvider stores the chosen provider (Itrocket, Krews, Jnode).
	selectedProvider string

	endpoints = config.DefaultEndpoints()
)

// snapshotCmd represents the main snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots for Story node",
	Long:  `Manage snapshots for Story node with subcommands like download and providers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If snapshot command is called by itself, guide the user
		cmd.Help()
		return nil
	},
}

func init() {
	// Attach subcommands to the main snapshot command
	snapshotCmd.AddCommand(downloadCmd)
	snapshotCmd.AddCommand(providersCmd)

	// Add flags to the download subcommand (e.g., home directory)
	downloadCmd.Flags().StringVar(&homeDirFlag, "home", defaultHomeDir(), "Home directory for Story node")

	// Flag to download snapshot directly to a specified path
	downloadCmd.Flags().String("output-path", "", "Download snapshot directly to the specified path without setup")
}

// defaultHomeDir returns the default home directory path
func defaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		pterm.Error.Println(fmt.Sprintf("Failed to get user home directory: %v", err))
		os.Exit(1)
	}
	return filepath.Join(home)
}

// GetSnapshotCmd returns the main snapshot command so it can be added
// to the root command in your main package.
func GetSnapshotCmd() *cobra.Command {
	return snapshotCmd
}
