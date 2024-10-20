package cmd

import (
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install story node",
}

func init() {
	rootCmd.AddCommand(setupCmd)
	// Add the node subcommand
	setupCmd.AddCommand(setupNodeCmd)
}
