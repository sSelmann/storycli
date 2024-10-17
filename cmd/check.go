// cmd/check.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// checkCmd represents the setup command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check story and geth services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
