package itrocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"

	"github.com/sSelmann/storycli/utils/bash"
	"github.com/sSelmann/storycli/utils/config"
	"github.com/sSelmann/storycli/utils/file"
)

var (
	// bestPrunedServerURL and bestArchiveServerURL are used to keep track
	// of the best server URLs for Itrocket (pruned vs. archive).
	bestPrunedServerURL  string
	bestArchiveServerURL string
)

type ItrocketSnapshotState struct {
	SnapshotName      string `json:"snapshot_name"`
	SnapshotGethName  string `json:"snapshot_geth_name"`
	SnapshotHeight    string `json:"snapshot_height"`
	SnapshotSize      string `json:"snapshot_size"`
	GethSnapshotSize  string `json:"geth_snapshot_size"`
	SnapshotBlockTime string `json:"snapshot_block_time"`
}

type itrocketServerData struct {
	state     ItrocketSnapshotState
	serverURL string
}

func DownloadSnapshotItrocket(homeDir, mode string) error {
	var serverURL string
	if strings.ToLower(mode) == "pruned" {
		serverURL = bestPrunedServerURL
	} else {
		serverURL = bestArchiveServerURL
	}
	if serverURL == "" {
		return errors.New("no best server URL found for the selected mode")
	}

	pterm.Info.Println(fmt.Sprintf("Fetching snapshot data from Itrocket (%s)...", serverURL))
	resp, err := http.Get(serverURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var snapshotState ItrocketSnapshotState
	if err := json.NewDecoder(resp.Body).Decode(&snapshotState); err != nil {
		return err
	}
	if snapshotState.SnapshotName == "" || snapshotState.SnapshotGethName == "" {
		return errors.New("failed to fetch snapshot names from Itrocket")
	}

	pterm.Info.Println("Installing required packages for Itrocket snapshot...")
	if err := bash.RunCommand("sudo", "apt", "install", "curl", "tmux", "jq", "lz4", "unzip", "-y"); err != nil {
		return err
	}

	pterm.Info.Println("Stopping Story and Story-Geth services...")
	if err := bash.RunCommand("sudo", "systemctl", "stop", "story", "story-geth"); err != nil {
		return err
	}

	pterm.Info.Println("Backup priv_validator_state.json...")
	if err := bash.RunCommand("cp", homeDir+"/.story/story/data/priv_validator_state.json", homeDir+"/.story/story/priv_validator_state.json.backup"); err != nil {
		return err
	}

	pterm.Info.Println("Removing old Story data...")
	storySnapshotURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(serverURL, "/.current_state.json"), snapshotState.SnapshotName)
	storySnapshotPath := filepath.Join(homeDir, ".story", "story_snapshot.tar.lz4")

	pterm.Info.Println("Downloading Story snapshot...")
	if err := file.DownloadFileWithFilteredProgress(storySnapshotURL, storySnapshotPath); err != nil {
		return err
	}

	pterm.Info.Println("Extracting Story snapshot...")
	if err := file.DecompressAndExtractLz4Tar(storySnapshotPath, filepath.Join(homeDir, ".story", "story")); err != nil {
		return err
	}
	if err := os.Remove(storySnapshotPath); err != nil {
		return err
	}

	pterm.Info.Println("Removing old Geth data...")
	gethSnapshotURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(serverURL, "/.current_state.json"), snapshotState.SnapshotGethName)
	gethSnapshotPath := filepath.Join(homeDir, ".story", "geth_snapshot.tar.lz4")

	pterm.Info.Println("Downloading Geth snapshot...")
	if err := file.DownloadFileWithFilteredProgress(gethSnapshotURL, gethSnapshotPath); err != nil {
		return err
	}
	pterm.Info.Println("Extracting Geth snapshot...")
	if err := file.DecompressAndExtractLz4Tar(gethSnapshotPath, filepath.Join(homeDir, ".story", "geth", "iliad", "geth")); err != nil {
		return err
	}
	if err := os.Remove(gethSnapshotPath); err != nil {
		return err
	}

	pterm.Info.Println("Restoring priv_validator_state.json...")
	if err := bash.RunCommand("mv", homeDir+"/.story/story/priv_validator_state.json.backup", homeDir+"/.story/story/data/priv_validator_state.json"); err != nil {
		return err
	}

	pterm.Info.Println("Starting Story and Story-Geth services...")
	if err := bash.RunCommand("sudo", "systemctl", "restart", "story", "story-geth"); err != nil {
		return err
	}

	pterm.Success.Println("Snapshot successfully downloaded and applied from Itrocket.")
	return nil
}

func DownloadSnapshotToPathItrocket(mode, path string, endpoint config.ItrocketEndpoints) error {
	var urls []string
	if mode == "pruned" {
		urls = endpoint.Pruned
	} else {
		urls = endpoint.Archive
	}

	best, err := fetchItrocketBestSnapshot(urls)
	if err != nil {
		return fmt.Errorf("failed to fetch best Itrocket snapshot: %v", err)
	}

	// Build full Story and Geth snapshot URLs
	storySnapshotURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(best.serverURL, "/.current_state.json"), best.state.SnapshotName)
	gethSnapshotURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(best.serverURL, "/.current_state.json"), best.state.SnapshotGethName)

	storyFileName := best.state.SnapshotName
	gethFileName := best.state.SnapshotGethName

	storyDestPath := filepath.Join(path, storyFileName)
	gethDestPath := filepath.Join(path, gethFileName)

	pterm.Info.Println(fmt.Sprintf("Downloading Itrocket Story snapshot from %s to %s...", storySnapshotURL, storyDestPath))
	err = file.DownloadFileWithFilteredProgress(storySnapshotURL, storyDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Itrocket Story snapshot: %v", err)
	}

	pterm.Info.Println(fmt.Sprintf("Downloading Itrocket Geth snapshot from %s to %s...", gethSnapshotURL, gethDestPath))
	err = file.DownloadFileWithFilteredProgress(gethSnapshotURL, gethDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Itrocket Geth snapshot: %v", err)
	}

	return nil
}

// fetchItrocketForMode fetches Itrocket data based on pruning mode
func FetchItrocketForMode(mode string, endpoint config.ItrocketEndpoints) (string, string, string, error) {
	var urls []string
	if mode == "pruned" {
		urls = endpoint.Pruned
	} else {
		urls = endpoint.Archive
	}

	best, err := fetchItrocketBestSnapshot(urls)
	if err != nil {
		return "", "", "", err
	}

	// Store the best server URL based on the pruning mode
	if mode == "pruned" {
		bestPrunedServerURL = best.serverURL
	} else {
		bestArchiveServerURL = best.serverURL
	}

	// Sum snapshot_size + geth_snapshot_size
	totalSize := sumSnapshotSizes(best.state.SnapshotSize, best.state.GethSnapshotSize)
	blockHeight := best.state.SnapshotHeight
	timeAgo := timeDifferenceString(best.state.SnapshotBlockTime)

	return totalSize, blockHeight, timeAgo, nil
}

// fetchItrocketBestSnapshot fetches the best (latest) snapshot from the given URLs
func fetchItrocketBestSnapshot(urls []string) (*itrocketServerData, error) {
	var bestData *itrocketServerData
	const layout = "2006-01-02T15:04:05.999999999Z07"

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			pterm.Warning.Printf("Could not fetch from %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		var state ItrocketSnapshotState
		if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
			pterm.Warning.Printf("Could not decode JSON from %s: %v\n", url, err)
			continue
		}

		bt, err := time.Parse(layout, state.SnapshotBlockTime)
		if err != nil {
			pterm.Warning.Printf("Could not parse snapshot_block_time from %s: %v\n", url, err)
			continue
		}

		currentData := &itrocketServerData{
			state:     state,
			serverURL: url,
		}
		if bestData == nil {
			bestData = currentData
		} else {
			bestBT, err2 := time.Parse(layout, bestData.state.SnapshotBlockTime)
			if err2 != nil {
				bestData = currentData
				continue
			}
			if bt.After(bestBT) {
				bestData = currentData
			}
		}
	}
	if bestData == nil {
		return nil, errors.New("no valid snapshot data found")
	}
	return bestData, nil
}

// sumSnapshotSizes merges e.g. "8.2G" + "44G" => "52.2G"
func sumSnapshotSizes(mainSize, gethSize string) string {
	val1 := parseAndConvertToFloat(mainSize)
	val2 := parseAndConvertToFloat(gethSize)
	if val1 < 0 || val2 < 0 {
		return "unknown"
	}
	total := val1 + val2
	return fmt.Sprintf("%.2fG", total)
}

func parseAndConvertToFloat(sizeStr string) float64 {
	if sizeStr == "" {
		return -1
	}
	s := strings.ToUpper(sizeStr)
	if strings.HasSuffix(s, "G") {
		s = strings.TrimSuffix(s, "G")
	} else if strings.HasSuffix(s, "GB") {
		s = strings.TrimSuffix(s, "GB")
	} else {
		return -1
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return val
}

func timeDifferenceString(timestr string) string {
	const layout = "2006-01-02T15:04:05.999999999Z07"
	t, err := time.Parse(layout, timestr)
	if err != nil {
		return "N/A"
	}
	duration := time.Since(t)
	if duration < 0 {
		return "N/A"
	}
	hours := int(duration.Hours())
	mins := int(duration.Minutes()) % 60
	if hours == 0 && mins == 0 {
		return "just now"
	} else if hours == 0 {
		return fmt.Sprintf("%dm ago", mins)
	} else {
		return fmt.Sprintf("%dh %dm ago", hours, mins)
	}
}
