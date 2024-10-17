// cmd/show.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show logs or configurations of story and geth",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
