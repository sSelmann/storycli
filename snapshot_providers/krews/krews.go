package krews

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/utils/bash"
)

type SnapshotKrews struct {
	Name         string      `json:"name"`
	Pruned       bool        `json:"pruned"`
	Size         string      `json:"size"`
	ChainID      string      `json:"chainId"`
	Block        json.Number `json:"block"`
	SnapshotDate string      `json:"snapshot_date"`
}

type KrewsSnapshotResponse struct {
	Snapshots []SnapshotKrews `json:"details"`
}

func DownloadSnapshotKrews(homeDir, pruningMode string) error {
	snapshotName := fmt.Sprintf("story_testnet_%s_snapshot", pruningMode)
	snapshotURL := fmt.Sprintf("krews-snapshot:krews-1-eu/%s", snapshotName)
	destDir := filepath.Join(homeDir, ".story")

	pterm.Info.Println("Installing and configuring Rclone for Krews snapshot...")
	err := installAndConfigureRcloneKrews(homeDir)
	if err != nil {
		return err
	}

	pterm.Info.Println("Backup priv_validator_state.json...")
	err = bash.RunCommand("cp", homeDir+"/.story/story/data/priv_validator_state.json", homeDir+"/.story/story/priv_validator_state.json.backup")
	if err != nil {
		return err
	}

	cmd := exec.Command("rclone", "copy", "--no-check-certificate", "--transfers=6", "--checkers=6", snapshotURL, destDir, "--progress")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	pterm.Info.Println("Restoring priv_validator_state.json...")
	err = bash.RunCommand("mv", homeDir+"/.story/story/priv_validator_state.json.backup"+homeDir+"/.story/story/data/priv_validator_state.json")
	if err != nil {
		return err
	}

	pterm.Success.Println("Snapshot successfully downloaded from Krews.")
	return nil
}

func DownloadSnapshotToPathKrews(mode, path string) error {
	snapshotName := fmt.Sprintf("story_testnet_%s_snapshot", mode)
	snapshotURL := fmt.Sprintf("krews-snapshot:krews-1-eu/%s", snapshotName)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pterm.Info.Println("Installing and configuring Rclone for Krews snapshot...")
	err = installAndConfigureRcloneKrews(homeDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("rclone", "copy", "--no-check-certificate", "--transfers=6", "--checkers=6", snapshotURL, path+"/"+snapshotName, "--progress")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("rclone copy failed: %v", err)
	}

	return nil
}

func FetchSnapshotSizesKrews(endpoint string) (
	prunedSize, archiveSize string,
	prunedBlockHeight, archiveBlockHeight string,
	prunedTimeAgo, archiveTimeAgo string,
	err error,
) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	defer resp.Body.Close()

	var snapshotResp KrewsSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return "", "", "", "", "", "", err
	}

	for _, snapshot := range snapshotResp.Snapshots {
		sz := snapshot.Size
		blk := snapshot.Block.String()
		timeAgo := parseKrewsSnapshotDate(snapshot.SnapshotDate)

		if snapshot.Pruned {
			prunedSize = sz
			prunedBlockHeight = blk
			prunedTimeAgo = timeAgo
		} else {
			archiveSize = sz
			archiveBlockHeight = blk
			archiveTimeAgo = timeAgo
		}
	}

	if prunedSize == "" {
		prunedSize = "unknown"
	}
	if archiveSize == "" {
		archiveSize = "unknown"
	}
	return prunedSize, archiveSize, prunedBlockHeight, archiveBlockHeight, prunedTimeAgo, archiveTimeAgo, nil
}

// parseKrewsSnapshotDate parses date strings like "26 Dec 2024, 18:17:50"
func parseKrewsSnapshotDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}
	const krewsLayout = "02 Jan 2006, 15:04:05"

	t, err := time.Parse(krewsLayout, dateStr)
	if err != nil {
		pterm.Warning.Printf("Could not parse Krews snapshot_date=%s: %v\n", dateStr, err)
		return "N/A"
	}
	duration := time.Since(t)
	if duration < 0 {
		return "N/A"
	}

	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60
	if h == 0 && m == 0 {
		return "just now"
	} else if h == 0 {
		return fmt.Sprintf("%dm ago", m)
	} else {
		return fmt.Sprintf("%dh %dm ago", h, m)
	}
}

func installAndConfigureRcloneKrews(homeDir string) error {
	// Check if Rclone is installed
	_, err := exec.LookPath("rclone")
	if err != nil {
		pterm.Info.Println("Rclone not found. Installing Rclone...")
		// Install Rclone
		err = bash.RunCommand("bash", "-c", "sudo -v; curl https://rclone.org/install.sh | sudo bash")
		if err != nil {
			return err
		}
		pterm.Success.Println("Rclone Installed")
	} else {
		pterm.Info.Println("Rclone is Already Installed")
	}

	// Configure Rclone for Krews
	pterm.Info.Println("Configuring Rclone for Krews...")
	rcloneConf := `[krews-snapshot]
type = s3
provider = DigitalOcean
region = fra1
endpoint = https://fra1.cdn.digitaloceanspaces.com
`
	rcloneConfDir := fmt.Sprintf("%s/.config/rclone", homeDir)
	err = os.MkdirAll(rcloneConfDir, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s/rclone.conf", rcloneConfDir), []byte(rcloneConf), 0644)
	if err != nil {
		return err
	}

	return nil
}
