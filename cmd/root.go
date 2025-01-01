// cmd/root.go
package cmd

import (
	"os"

	"github.com/sSelmann/storycli/cmd/snapshot"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "storycli",
	Short: "Story CLI is built for setting up and managing your Story node.",
	Long: `Story CLI is built for setting up and managing your Story node.
It has commands to make your work easier and save your time.`,
}

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	completion := completionCommand()

	// mark completion hidden
	completion.Hidden = true
	rootCmd.AddCommand(completion)

	rootCmd.AddCommand(snapshot.GetSnapshotCmd())

	rootCmd.Flags().BoolP("help", "h", false, "help for storycli")
}
