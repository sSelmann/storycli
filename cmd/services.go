// cmd/services.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage story and geth services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("services called")
	},
}

func init() {
	rootCmd.AddCommand(servicesCmd)
}
