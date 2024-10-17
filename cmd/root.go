// cmd/root.go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "storycli",
	Short: "Story CLI is built for setting up and managing your Story node.",
	Long: `Story CLI is built for setting up and managing your Story node.
It has commands to make your work easier and save your time.`,

	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("help", "h", false, "help for storycli")
}

