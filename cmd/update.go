// cmd/update.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		printInfo("Updating geth...")
		err := runCommand("cd $HOME && wget -O geth https://github.com/piplabs/story-geth/releases/latest/download/geth-linux-amd64")
		if err != nil {
			printError("Failed to download geth")
		}
		err = runCommand("chmod +x $HOME/geth")
		if err != nil {
			printError("Failed to make geth executable")
		}
		err = runCommand("sudo mv $HOME/geth $(which geth)")
		if err != nil {
			printError("Failed to move geth binary")
		}
		err = runCommand("sudo systemctl restart story-geth")
		if err != nil {
			printError("Failed to restart story-geth")
		}
		err = runCommand("sudo systemctl restart story && sudo journalctl -u story -f")
		if err != nil {
			printError("Failed to restart story")
		}

		// Update story binary
		printInfo("Updating story...")
		err = runCommand("cd $HOME && rm -rf story")
		if err != nil {
			printError("Failed to remove old story")
		}
		err = runCommand("git clone https://github.com/piplabs/story")
		if err != nil {
			printError("Failed to clone story repository")
		}

		// Get the latest release tag
		tag, err := getLatestReleaseTag("piplabs/story")
		if err != nil {
			printError("Failed to get latest version")
		}

		err = runCommand(fmt.Sprintf("cd story && git checkout %s", tag))
		if err != nil {
			printError("Failed to checkout story version")
		}
		err = runCommand("cd $HOME/story && go build -o story ./client")
		if err != nil {
			printError("Failed to build story")
		}
		err = runCommand("mv $HOME/story/story $HOME/go/bin/")
		if err != nil {
			printError("Failed to move story binary")
		}

		fmt.Println("Update completed successfully.")
	},
}
