package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Story and Story-Geth services",
	RunE:  runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
	printInfo("Checking Story and Geth services exist...")

	storyExists, err := checkServiceExists("story")
	if err != nil {
		return err
	}

	gethExists, err := checkServiceExists("story-geth")
	if err != nil {
		return err
	}

	if !storyExists && !gethExists {
		printWarning("Neither 'story' nor 'story-geth' services are installed.")
		return nil
	}

	if storyExists {
		printInfo("Restarting story service...")
		err = displayRestart("story")
		if err != nil {

		}
	} else {
		printWarning("'story' service is not installed.")
	}

	if gethExists {
		printInfo("Restarting story-geth service...")
		err = displayRestart("story-geth")
		if err != nil {
		}
	} else {
		printWarning("'story-geth' service is not installed.")
	}

	printSuccess("Story and Geth services successfully restarted")
	return nil
}

func displayRestart(serviceName string) error {
	cmdStr := fmt.Sprintf("systemctl restart %s", serviceName)
	cmd := exec.Command("bash", "-c", cmdStr)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Servis durumu ne olursa olsun çıktıyı göster
	if out.Len() > 0 {
		fmt.Println(out.String())
	}
	if stderr.Len() > 0 {
		fmt.Println("Stderr:", stderr.String())
	}

	// Hata kontrolü
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil
		} else {
			// Diğer türde hatalar için uyarı göster
			printWarning(fmt.Sprintf("Failed to get status for %s service.", serviceName))
			return err
		}
	}

	return nil
}
