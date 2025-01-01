package jnode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/utils/bash"
	"github.com/sSelmann/storycli/utils/config"
	"github.com/sSelmann/storycli/utils/file"
)

var endpoints config.Endpoints

// JnodeSnapshotFiles represents the files section in jnode API response
type JnodeSnapshotFiles struct {
	Geth  JnodeSnapshotFile `json:"geth"`
	Story JnodeSnapshotFile `json:"story"`
}

// JnodeSnapshotFile represents individual file information
type JnodeSnapshotFile struct {
	SizeGB float64 `json:"size_gb"`
	URL    string  `json:"url"`
}

// JnodeSnapshotMode represents pruned or archive data
type JnodeSnapshotMode struct {
	Files          JnodeSnapshotFiles `json:"files"`
	SnapshotHeight string             `json:"snapshot_height"`
	TimeAgo        string             `json:"time_ago"`
}

// JnodeSnapshotResponse represents the entire API response from jnode
type JnodeSnapshotResponse struct {
	Archive JnodeSnapshotMode `json:"archive"`
	Pruned  JnodeSnapshotMode `json:"pruned"`
}

// FetchSnapshotSizesJnode fetches snapshot sizes and details from jnode API
func FetchSnapshotSizesJnode() (
	prunedSize, archiveSize string,
	prunedBlockHeight, archiveBlockHeight string,
	prunedTimeAgo, archiveTimeAgo string,
	err error,
) {
	apiURL := "https://snapshot-external-providers-api.krews.xyz/snapshots/jnode"

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", "", "", fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	var snapshotResp JnodeSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return "", "", "", "", "", "", err
	}

	// PRUNED
	prunedStoryGB := snapshotResp.Pruned.Files.Story.SizeGB
	prunedGethGB := snapshotResp.Pruned.Files.Geth.SizeGB
	prunedSumGB := prunedStoryGB + prunedGethGB // Combine
	prunedSize = fmt.Sprintf("%.2fG", prunedSumGB)
	prunedBlockHeight = snapshotResp.Pruned.SnapshotHeight
	prunedTimeAgo = snapshotResp.Pruned.TimeAgo

	// ARCHIVE
	archiveStoryGB := snapshotResp.Archive.Files.Story.SizeGB
	archiveGethGB := snapshotResp.Archive.Files.Geth.SizeGB
	archiveSumGB := archiveStoryGB + archiveGethGB
	archiveSize = fmt.Sprintf("%.2fG", archiveSumGB)
	archiveBlockHeight = snapshotResp.Archive.SnapshotHeight
	archiveTimeAgo = snapshotResp.Archive.TimeAgo

	return prunedSize, archiveSize,
		prunedBlockHeight, archiveBlockHeight,
		prunedTimeAgo, archiveTimeAgo,
		nil
}

// DownloadSnapshotToPathJnode downloads the Jnode snapshot to a specified path without applying it
func DownloadSnapshotToPathJnode(mode, path string) error {
	apiURL := endpoints.Jnode

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Jnode snapshot data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status from Jnode API: %s", resp.Status)
	}

	var snapshotResp JnodeSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return fmt.Errorf("failed to decode Jnode API response: %v", err)
	}

	var storySnapshotURL, gethSnapshotURL string
	if mode == "pruned" {
		storySnapshotURL = snapshotResp.Pruned.Files.Story.URL
		gethSnapshotURL = snapshotResp.Pruned.Files.Geth.URL
	} else if mode == "archive" {
		storySnapshotURL = snapshotResp.Archive.Files.Story.URL
		gethSnapshotURL = snapshotResp.Archive.Files.Geth.URL
	} else {
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	storyFileName := filepath.Base(storySnapshotURL)
	gethFileName := filepath.Base(gethSnapshotURL)

	storyDestPath := filepath.Join(path, storyFileName)
	gethDestPath := filepath.Join(path, gethFileName)

	pterm.Info.Printf(fmt.Sprintf("Downloading Jnode Story snapshot from %s to %s...", storySnapshotURL, storyDestPath))
	err = file.DownloadFileWithProgress(storySnapshotURL, storyDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Jnode Story snapshot: %v", err)
	}

	pterm.Info.Printf(fmt.Sprintf("Downloading Jnode Geth snapshot from %s to %s...", gethSnapshotURL, gethDestPath))
	err = file.DownloadFileWithProgress(gethSnapshotURL, gethDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Jnode Geth snapshot: %v", err)
	}

	return nil
}

// DownloadSnapshotJnode downloads and applies the Jnode snapshot
func DownloadSnapshotJnode(homeDir, mode string) error {
	pterm.Info.Printf("Installing required packages for Jnode snapshot...")
	if err := bash.RunCommand("sudo apt-get install wget lz4 aria2 pv -y"); err != nil {
		return err
	}

	pterm.Info.Printf("Stopping Story and Story-Geth services...")
	if err := bash.RunCommand("sudo systemctl stop story story-geth"); err != nil {
		return err
	}

	pterm.Info.Printf("Backing up priv_validator_state.json...")
	if err := bash.RunCommand(fmt.Sprintf("cp %s/.story/story/data/priv_validator_state.json %s/.story/priv_validator_state.json.backup", homeDir, homeDir)); err != nil {
		return err
	}

	pterm.Info.Printf("Removing old Story and Geth data...")
	if err := bash.RunCommand(fmt.Sprintf("rm -f %s/.story/story/data/*", homeDir)); err != nil {
		return err
	}
	if err := bash.RunCommand(fmt.Sprintf("rm -rf %s/.story/geth/odyssey/geth/chaindata", homeDir)); err != nil {
		return err
	}

	// Fetch snapshot URLs from the Jnode API
	apiURL := endpoints.Jnode
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Jnode snapshot data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status from Jnode API: %s", resp.Status)
	}

	var snapshotResp JnodeSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return fmt.Errorf("failed to decode Jnode API response: %v", err)
	}

	var storySnapshotURL, gethSnapshotURL string
	if mode == "pruned" {
		storySnapshotURL = snapshotResp.Pruned.Files.Story.URL
		gethSnapshotURL = snapshotResp.Pruned.Files.Geth.URL
	} else if mode == "archive" {
		storySnapshotURL = snapshotResp.Archive.Files.Story.URL
		gethSnapshotURL = snapshotResp.Archive.Files.Geth.URL
	} else {
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	pterm.Info.Printf("Downloading Story snapshot...")
	storySnapshotPath := filepath.Join(homeDir, "Story_snapshot.lz4")
	if err := bash.RunCommand(fmt.Sprintf("aria2c -x 16 -s 16 -k 1M %s -o %s", storySnapshotURL, storySnapshotPath)); err != nil {
		return fmt.Errorf("failed to download Story snapshot: %v", err)
	}

	pterm.Info.Printf("Downloading Geth snapshot...")
	gethSnapshotPath := filepath.Join(homeDir, "Geth_snapshot.lz4")
	if err := bash.RunCommand(fmt.Sprintf("aria2c -x 16 -s 16 -k 1M %s -o %s", gethSnapshotURL, gethSnapshotPath)); err != nil {
		return fmt.Errorf("failed to download Geth snapshot: %v", err)
	}

	pterm.Info.Printf("Extracting Story snapshot...")
	if err := bash.RunCommand(fmt.Sprintf("lz4 -d -c %s | pv | sudo tar xv -C %s/.story/story/ > /dev/null", storySnapshotPath, homeDir)); err != nil {
		return fmt.Errorf("failed to extract Story snapshot: %v", err)
	}
	if err := bash.RunCommand(fmt.Sprintf("rm -f %s", storySnapshotPath)); err != nil {
		return err
	}

	pterm.Info.Printf("Extracting Geth snapshot...")
	if err := bash.RunCommand(fmt.Sprintf("lz4 -d -c %s | pv | sudo tar xv -C %s/.story/geth/odyssey/geth/ > /dev/null", gethSnapshotPath, homeDir)); err != nil {
		return fmt.Errorf("failed to extract Geth snapshot: %v", err)
	}
	if err := bash.RunCommand(fmt.Sprintf("rm -f %s", gethSnapshotPath)); err != nil {
		return err
	}

	pterm.Info.Printf("Restoring priv_validator_state.json...")
	if err := bash.RunCommand(fmt.Sprintf("cp %s/.story/priv_validator_state.json.backup %s/.story/story/data/priv_validator_state.json", homeDir, homeDir)); err != nil {
		return err
	}

	pterm.Info.Printf("Starting Story and Story-Geth services...")
	if err := bash.RunCommand("sudo systemctl restart story story-geth"); err != nil {
		return err
	}

	pterm.Success.Printf("Snapshot successfully downloaded and applied from Jnode.")
	return nil
}
