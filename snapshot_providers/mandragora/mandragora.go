package mandragora

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/utils/bash"
	"github.com/sSelmann/storycli/utils/file"
)

// MandragoraSnapshotFiles represents the files section in Mandragora API response
type MandragoraSnapshotFiles struct {
	Geth  MandragoraSnapshotFile `json:"geth"`
	Story MandragoraSnapshotFile `json:"story"`
}

// MandragoraSnapshotFile represents individual file information
type MandragoraSnapshotFile struct {
	Size string `json:"size"` // Correctly include the "size" field
	URL  string `json:"url"`
}

// MandragoraSnapshotMode represents pruned or archive data
type MandragoraSnapshotMode struct {
	Files          MandragoraSnapshotFiles `json:"files"`
	SnapshotHeight string                  `json:"snapshot_height"`
	TimeAgo        string                  `json:"time_ago"`
}

// MandragoraSnapshotResponse represents the entire API response from Mandragora
type MandragoraSnapshotResponse struct {
	Archive MandragoraSnapshotMode `json:"archive"`
	Pruned  MandragoraSnapshotMode `json:"pruned"`
}

// FetchSnapshotSizesMandragora fetches snapshot sizes and details from mandragora API
func FetchSnapshotSizesMandragora() (
	prunedSize, archiveSize string,
	prunedBlockHeight, archiveBlockHeight string,
	prunedTimeAgo, archiveTimeAgo string,
	err error,
) {
	apiURL := "https://snapshot-external-providers-api.krews.xyz/snapshots/mandragora"

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", "", "", fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	var snapshotResp MandragoraSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return "", "", "", "", "", "", err
	}

	// PRUNED
	prunedStorySize := parseSize(snapshotResp.Pruned.Files.Story.Size)
	prunedGethSize := parseSize(snapshotResp.Pruned.Files.Geth.Size)
	prunedSumSize := prunedStorySize + prunedGethSize
	if prunedSumSize == 0 {
		prunedSize = "0.00M"
	} else {
		prunedSize = fmt.Sprintf("%.2fM", prunedSumSize) // Display in MB
	}
	prunedBlockHeight = snapshotResp.Pruned.SnapshotHeight
	if prunedBlockHeight == "" {
		prunedBlockHeight = "unknown"
	}
	prunedTimeAgo = snapshotResp.Pruned.TimeAgo
	if prunedTimeAgo == "" {
		prunedTimeAgo = "unknown"
	} else {
		prunedTimeAgo = fmt.Sprintf("%s ago", prunedTimeAgo)
	}

	// ARCHIVE
	archiveStorySize := parseSize(snapshotResp.Archive.Files.Story.Size)
	archiveGethSize := parseSize(snapshotResp.Archive.Files.Geth.Size)
	archiveSumSize := archiveStorySize + archiveGethSize
	if archiveSumSize == 0 {
		archiveSize = "0.00G"
	} else {
		archiveSize = fmt.Sprintf("%.2fG", archiveSumSize/1024) // Convert MB to GB for display
	}
	archiveBlockHeight = snapshotResp.Archive.SnapshotHeight
	if archiveBlockHeight == "" {
		archiveBlockHeight = "unknown"
	}
	archiveTimeAgo = snapshotResp.Archive.TimeAgo
	if archiveTimeAgo == "" {
		archiveTimeAgo = "unknown"
	} else {
		archiveTimeAgo = fmt.Sprintf("%s ago", archiveTimeAgo)
	}

	return prunedSize, archiveSize,
		prunedBlockHeight, archiveBlockHeight,
		prunedTimeAgo, archiveTimeAgo,
		nil
}

// parseSize converts size strings (e.g., "55G", "1.2M", "36K") to GB
func parseSize(sizeStr string) float64 {
	if sizeStr == "" {
		return 0
	}

	sizeStr = strings.ToUpper(sizeStr) // Ensure consistency in unit parsing
	if strings.HasSuffix(sizeStr, "G") {
		size, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "G"), 64)
		return size * 1024 // Convert GB to MB
	} else if strings.HasSuffix(sizeStr, "M") {
		size, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "M"), 64)
		return size // Already in MB
	} else if strings.HasSuffix(sizeStr, "K") {
		size, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "K"), 64)
		return size / 1024 // Convert KB to MB
	}

	// If no unit is specified or unsupported unit
	size, _ := strconv.ParseFloat(sizeStr, 64)
	return size
}

// DownloadSnapshotToPathMandragora downloads the Mandragora snapshot to a specified path without applying it
func DownloadSnapshotToPathMandragora(mode, path string, endpoint string) error {
	apiURL := endpoint

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Mandragora snapshot data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status from Mandragora API: %s", resp.Status)
	}

	var snapshotResp MandragoraSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return fmt.Errorf("failed to decode Mandragora API response: %v", err)
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

	pterm.Info.Println(fmt.Sprintf("Downloading Mandragora Story snapshot from %s to %s...", storySnapshotURL, storyDestPath))
	err = file.DownloadFileWithAria2(storySnapshotURL, storyDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Mandragora Story snapshot: %v", err)
	}

	pterm.Info.Println(fmt.Sprintf("Downloading Mandragora Geth snapshot from %s to %s...", gethSnapshotURL, gethDestPath))
	err = file.DownloadFileWithAria2(gethSnapshotURL, gethDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Mandragora Geth snapshot: %v", err)
	}

	return nil
}

// DownloadSnapshotMandragora downloads and applies the Mandragora snapshot
func DownloadSnapshotMandragora(homeDir, mode string, endpoint string, isCosmovisor bool) error {
	pterm.Info.Println("Installing required packages for Mandragora snapshot...")
	if err := bash.RunCommand("sudo", "apt-get", "install", "wget", "lz4", "aria2", "pv", "-y"); err != nil {
		return err
	}

	pterm.Info.Println("Stopping Story and Story-Geth services...")
	if err := bash.RunCommand("sudo", "systemctl", "stop", "story", "story-geth"); err != nil {
		return err
	}

	pterm.Info.Println("Backing up priv_validator_state.json...")
	if err := bash.RunCommand("cp", homeDir+"/.story/story/data/priv_validator_state.json", homeDir+"/.story/priv_validator_state.json.backup"); err != nil {
		return err
	}

	pterm.Info.Println("Removing old Story and Geth data...")
	if err := bash.RunCommand("rm", "-rf", homeDir+"/.story/story/data/*"); err != nil {
		return err
	}
	if err := bash.RunCommand("rm", "-rf", homeDir+"/.story/geth/odyssey/geth/chaindata"); err != nil {
		return err
	}

	// Fetch snapshot URLs from the Mandragora API
	apiURL := endpoint
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Mandragora snapshot data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status from Mandragora API: %s", resp.Status)
	}

	var snapshotResp MandragoraSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshotResp); err != nil {
		return fmt.Errorf("failed to decode Mandragora API response: %v", err)
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

	pterm.Info.Println("Downloading Story snapshot...")
	storySnapshotPath := filepath.Join(homeDir, "Story_snapshot.lz4")
	file.DownloadFileWithAria2(storySnapshotURL, storySnapshotPath)

	pterm.Info.Println("Downloading Geth snapshot...")
	gethSnapshotPath := filepath.Join(homeDir, "Geth_snapshot.lz4")
	file.DownloadFileWithAria2(gethSnapshotURL, gethSnapshotPath)

	pterm.Info.Println("Extracting Story snapshot...")
	if err := file.DecompressAndExtractLz4Tar(storySnapshotPath, filepath.Join(homeDir, ".story", "story")); err != nil {
		return err
	}
	if err := bash.RunCommand("rm", "-rf", storySnapshotPath); err != nil {
		return err
	}

	pterm.Info.Println("Extracting Geth snapshot...")
	if err := file.DecompressAndExtractLz4Tar(gethSnapshotPath, filepath.Join(homeDir, ".story", "geth", "odyssey", "geth")); err != nil {
		return err
	}
	if err := bash.RunCommand("rm", "-rf", gethSnapshotPath); err != nil {
		return err
	}

	pterm.Info.Println("Restoring priv_validator_state.json...")
	if err := bash.RunCommand("cp", homeDir+"/.story/priv_validator_state.json.backup", homeDir+"/.story/story/data/priv_validator_state.json"); err != nil {
		return err
	}

	if !isCosmovisor {
		pterm.Info.Println("Starting Story and Story-Geth services...")
		if err := bash.RunCommand("sudo", "systemctl", "restart", "story", "story-geth"); err != nil {
			return err
		}
	}

	pterm.Success.Println("Snapshot successfully downloaded and applied from Mandragora.")
	return nil
}
