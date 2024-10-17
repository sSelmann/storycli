// cmd/snapshot.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Download a snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("snapshot called")
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
