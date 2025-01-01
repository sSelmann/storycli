// cmd/update.go
package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/sSelmann/storycli/utils/bash"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

// setupCmd represents the setup command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update story and geth binaries",
	Run: func(cmd *cobra.Command, args []string) {
		// Update geth binary
		pterm.Info.Printf("Updating geth...")
		err := bash.RunCommand("cd $HOME && wget -O geth https://github.com/piplabs/story-geth/releases/latest/download/geth-linux-amd64")
		if err != nil {
			printError("Failed to download geth")
		}
		err = bash.RunCommand("chmod +x $HOME/geth")
		if err != nil {
			printError("Failed to make geth executable")
		}
		err = bash.RunCommand("sudo mv $HOME/geth $(which geth)")
		if err != nil {
			printError("Failed to move geth binary")
		}
		err = bash.RunCommand("sudo systemctl restart story-geth")
		if err != nil {
			printError("Failed to restart story-geth")
		}
		err = bash.RunCommand("sudo systemctl restart story && sudo journalctl -u story -f")
		if err != nil {
			printError("Failed to restart story")
		}

		// Update story binary
		pterm.Info.Printf("Updating story...")
		err = bash.RunCommand("cd $HOME && rm -rf story")
		if err != nil {
			printError("Failed to remove old story")
		}
		err = bash.RunCommand("git clone https://github.com/piplabs/story")
		if err != nil {
			printError("Failed to clone story repository")
		}

		// Get the latest release tag
		tag, err := getLatestReleaseTag("piplabs/story")
		if err != nil {
			printError("Failed to get latest version")
		}

		err = bash.RunCommand(fmt.Sprintf("cd story && git checkout %s", tag))
		if err != nil {
			printError("Failed to checkout story version")
		}
		err = bash.RunCommand("cd $HOME/story && go build -o story ./client")
		if err != nil {
			printError("Failed to build story")
		}
		err = bash.RunCommand("mv $HOME/story/story $HOME/go/bin/")
		if err != nil {
			printError("Failed to move story binary")
		}

		fmt.Println("Update completed successfully.")
	},
}
