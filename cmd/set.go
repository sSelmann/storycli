// cmd/set.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configurations of story and geth",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("set called")
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
