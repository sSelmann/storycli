// cmd/version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Check versions of existing story and geth binaries",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version called")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
