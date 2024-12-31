package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/sSelmann/storycli/utils/config"
)

// -----------------------------------------------------------------------------
// Global Variables
// -----------------------------------------------------------------------------

var (
	// homeDirFlag is used to store the home directory path (passed via --home).
	homeDirFlag string

	// pruningMode represents the selected pruning mode ("pruned" or "archive").
	pruningMode string

	// bestPrunedServerURL and bestArchiveServerURL are used to keep track
	// of the best server URLs for Itrocket (pruned vs. archive).
	bestPrunedServerURL  string
	bestArchiveServerURL string

	// selectedProvider stores the chosen provider (Itrocket, Krews, Jnode).
	selectedProvider string

	// endpoints holds the dynamic endpoints (Itrocket, Krews, Jnode).
	endpoints config.Endpoints
)

// -----------------------------------------------------------------------------
// Cobra Commands
// -----------------------------------------------------------------------------

// snapshotCmd represents the main snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots for Story node",
	Long:  `Manage snapshots for Story node with subcommands like download and providers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If snapshot command is called by itself, guide the user
		cmd.Help()
		return nil
	},
}

// downloadCmd represents the snapshot download subcommand
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download and apply snapshots",
	Long:  `Download and apply snapshots from selected providers to initialize or update your Story node.`,
	RunE:  runDownloadSnapshot,
}

// providersCmd represents the snapshot providers subcommand
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available snapshot providers and their data",
	Long:  `List available snapshot providers and display their snapshot data in a table format.`,
	RunE:  runListProviders,
}

// -----------------------------------------------------------------------------
// init
// -----------------------------------------------------------------------------

func init() {
	// Attach subcommands to the main snapshot command
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(downloadCmd)
	snapshotCmd.AddCommand(providersCmd)

	// Add flags to the download subcommand (e.g., home directory)
	downloadCmd.Flags().StringVar(&homeDirFlag, "home", defaultHomeDir(), "Home directory for Story node")

	// Flag to download snapshot directly to a specified path
	downloadCmd.Flags().String("output-path", "", "Download snapshot directly to the specified path without setup")

	// Dynamically fetch endpoints from config
	endpoints = config.DefaultEndpoints()
}

// defaultHomeDir returns the default home directory path
func defaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		os.Exit(1)
	}
	return filepath.Join(home)
}

// -----------------------------------------------------------------------------
// "snapshot download" Subcommand
// -----------------------------------------------------------------------------

func runDownloadSnapshot(cmd *cobra.Command, args []string) error {
	// Retrieve the output-path flag if any
	outputPath, err := cmd.Flags().GetString("output-path")
	if err != nil {
		return fmt.Errorf("failed to parse output-path flag: %v", err)
	}

	// 1) Ask for pruning mode FIRST
	printInfo("Pruning Mode Information:")
	pterm.Info.Println(" - Pruned Mode: Stores only recent blocks, reducing disk usage.")
	pterm.Info.Println(" - Archive Mode: Stores the entire chain history, requiring more disk space.\n")

	if pruningMode == "" {
		pmPrompt := promptui.Select{
			Label: "Select the pruning mode",
			Items: []string{"pruned", "archive"},
		}
		_, pmResult, err := pmPrompt.Run()
		if err != nil {
			return err
		}
		pruningMode = pmResult
	}

	// 2) Fetch data for all providers, specifically for the chosen pruning mode
	printInfo(fmt.Sprintf("Fetching snapshot data for providers (mode=%s)...", pruningMode))
	providersData, err := fetchAllProvidersDataForMode(pruningMode) // includes Jnode
	if err != nil {
		printWarning(fmt.Sprintf("Failed to fetch providers data: %v", err))
	}

	if len(providersData) == 0 {
		return errors.New("no snapshot data found for any provider")
	}

	// 3) Show provider selection prompt
	selectedProvider, err = selectSnapshotProvider(providersData)
	if err != nil {
		return err
	}

	// If --output-path is set, download directly and skip extra steps
	if outputPath != "" {
		printInfo(fmt.Sprintf("Downloading snapshot from %s in %s mode to %s...", selectedProvider, pruningMode, outputPath))

		switch selectedProvider {
		case "Itrocket":
			err = downloadSnapshotToPathItrocket(pruningMode, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download snapshot from Itrocket: %v", err)
			}
		case "Krews":
			err = downloadSnapshotToPathKrews(pruningMode, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download snapshot from Krews: %v", err)
			}
		case "Jnode":
			err = downloadSnapshotToPathJnode(pruningMode, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download snapshot from Jnode: %v", err)
			}
		default:
			return errors.New("provider not supported yet")
		}

		printSuccess(fmt.Sprintf("Snapshot successfully downloaded to %s from %s.", outputPath, selectedProvider))
		return nil
	}

	// 4) Stop services
	printInfo("Stopping services...")
	if err := runCommand("sudo systemctl stop story story-geth"); err != nil {
		return err
	}

	// 5) Download the snapshot from chosen provider & mode
	switch selectedProvider {
	case "Itrocket":
		if err := downloadSnapshotItrocket(homeDirFlag, pruningMode); err != nil {
			return err
		}
	case "Krews":
		if err := downloadSnapshotKrews(homeDirFlag, pruningMode); err != nil {
			return err
		}
	case "Jnode":
		if err := downloadSnapshotJnode(homeDirFlag, pruningMode); err != nil {
			return err
		}
	default:
		return errors.New("provider not supported yet")
	}

	// 6) Prompt user to restart
	pterm.Info.Printf(
		"Run %s to restart the systemd services and apply the new values\n",
		pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprintf("scli restart"),
	)

	return nil
}

// -----------------------------------------------------------------------------
// "snapshot providers" Subcommand
// -----------------------------------------------------------------------------

func runListProviders(cmd *cobra.Command, args []string) error {
	// We'll fetch data for both pruned and archive
	modes := []string{"pruned", "archive"}

	providersData, err := fetchAllProvidersDataForModes(modes)
	if err != nil {
		return fmt.Errorf("failed to fetch providers data: %v", err)
	}

	if len(providersData) == 0 {
		return errors.New("no providers data available")
	}

	// Separate data into pruned vs. archive
	prunedProviders := []providerSnapshotInfo{}
	archiveProviders := []providerSnapshotInfo{}

	for _, pd := range providersData {
		if pd.Mode == "pruned" {
			prunedProviders = append(prunedProviders, pd)
		} else if pd.Mode == "archive" {
			archiveProviders = append(archiveProviders, pd)
		}
	}

	// Create a table for pruned snapshots
	if len(prunedProviders) > 0 {
		pterm.DefaultSection.Println("Pruned Snapshots")
		prunedTableData := pterm.TableData{
			{"Provider", "Total Size", "Block Height", "Time Ago"},
		}

		for _, pd := range prunedProviders {
			prunedTableData = append(prunedTableData, []string{
				pd.ProviderName,
				pd.TotalSize,
				pd.BlockHeight,
				pd.TimeAgo,
			})
		}

		err = pterm.DefaultTable.
			WithHasHeader(true).
			WithData(prunedTableData).
			Render()
		if err != nil {
			return fmt.Errorf("failed to render pruned table: %v", err)
		}
	}

	// Create a table for archive snapshots
	if len(archiveProviders) > 0 {
		pterm.DefaultSection.Println("Archive Snapshots")
		archiveTableData := pterm.TableData{
			{"Provider", "Total Size", "Block Height", "Time Ago"},
		}

		for _, pd := range archiveProviders {
			archiveTableData = append(archiveTableData, []string{
				pd.ProviderName,
				pd.TotalSize,
				pd.BlockHeight,
				pd.TimeAgo,
			})
		}

		err = pterm.DefaultTable.
			WithHasHeader(true).
			WithData(archiveTableData).
			Render()
		if err != nil {
			return fmt.Errorf("failed to render archive table: %v", err)
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Provider Snapshot Info
// -----------------------------------------------------------------------------

// providerSnapshotInfo holds data displayed for each provider
type providerSnapshotInfo struct {
	ProviderName string
	Mode         string
	TotalSize    string // sum of snapshot_size + geth_snapshot_size (for Itrocket)
	BlockHeight  string
	TimeAgo      string
}

// -----------------------------------------------------------------------------
// Fetching Data for Multiple Pruning Modes
// -----------------------------------------------------------------------------

func fetchAllProvidersDataForModes(modes []string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	for _, mode := range modes {
		// ITROCKET
		totalSizeIt, blockHeightIt, timeAgoIt, err := fetchItrocketForMode(mode)
		if err != nil {
			pterm.Warning.Printf("Failed to fetch Itrocket data (mode=%s): %v\n", mode, err)
			results = append(results, providerSnapshotInfo{
				ProviderName: "Itrocket",
				Mode:         mode,
				TotalSize:    "unknown",
				BlockHeight:  "N/A",
				TimeAgo:      "N/A",
			})
		} else {
			results = append(results, providerSnapshotInfo{
				ProviderName: "Itrocket",
				Mode:         mode,
				TotalSize:    totalSizeIt,
				BlockHeight:  blockHeightIt,
				TimeAgo:      timeAgoIt,
			})
		}

		// KREWS
		kPrunedSize, kArchiveSize,
			kPrunedBlock, kArchiveBlock,
			kPrunedTimeAgo, kArchiveTimeAgo,
			err := fetchSnapshotSizesKrews()
		if err != nil {
			pterm.Warning.Printf("Failed to fetch Krews data (mode=%s): %v\n", mode, err)
			results = append(results, providerSnapshotInfo{
				ProviderName: "Krews",
				Mode:         mode,
				TotalSize:    "unknown",
				BlockHeight:  "N/A",
				TimeAgo:      "N/A",
			})
		} else {
			var size, blockH, timeAgo string
			if mode == "pruned" {
				size = kPrunedSize
				blockH = kPrunedBlock
				timeAgo = kPrunedTimeAgo
			} else {
				size = kArchiveSize
				blockH = kArchiveBlock
				timeAgo = kArchiveTimeAgo
			}
			results = append(results, providerSnapshotInfo{
				ProviderName: "Krews",
				Mode:         mode,
				TotalSize:    size,
				BlockHeight:  blockH,
				TimeAgo:      timeAgo,
			})
		}

		// JNODE
		jPrunedSize, jArchiveSize,
			jPrunedBlock, jArchiveBlock,
			jPrunedTimeAgo, jArchiveTimeAgo,
			err := fetchSnapshotSizesJnode()
		if err != nil {
			pterm.Warning.Printf("Failed to fetch Jnode data (mode=%s): %v\n", mode, err)
			results = append(results, providerSnapshotInfo{
				ProviderName: "Jnode",
				Mode:         mode,
				TotalSize:    "unknown",
				BlockHeight:  "N/A",
				TimeAgo:      "N/A",
			})
		} else {
			var size, blockH, timeAgo string
			if mode == "pruned" {
				size = jPrunedSize
				blockH = jPrunedBlock
				timeAgo = jPrunedTimeAgo
			} else {
				size = jArchiveSize
				blockH = jArchiveBlock
				timeAgo = jArchiveTimeAgo
			}
			results = append(results, providerSnapshotInfo{
				ProviderName: "Jnode",
				Mode:         mode,
				TotalSize:    size,
				BlockHeight:  blockH,
				TimeAgo:      timeAgo,
			})
		}
	}

	return results, nil
}

// -----------------------------------------------------------------------------
// Fetching Data for a Single Pruning Mode (Used by runDownloadSnapshot)
// -----------------------------------------------------------------------------

func fetchAllProvidersDataForMode(mode string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	// ITROCKET
	totalSizeIt, blockHeightIt, timeAgoIt, err := fetchItrocketForMode(mode)
	if err != nil {
		pterm.Warning.Printf("Failed to fetch Itrocket data (mode=%s): %v\n", mode, err)
		results = append(results, providerSnapshotInfo{
			ProviderName: "Itrocket",
			Mode:         mode,
			TotalSize:    "unknown",
			BlockHeight:  "N/A",
			TimeAgo:      "N/A",
		})
	} else {
		results = append(results, providerSnapshotInfo{
			ProviderName: "Itrocket",
			Mode:         mode,
			TotalSize:    totalSizeIt,
			BlockHeight:  blockHeightIt,
			TimeAgo:      timeAgoIt,
		})
	}

	// KREWS
	kPrunedSize, kArchiveSize,
		kPrunedBlock, kArchiveBlock,
		kPrunedTimeAgo, kArchiveTimeAgo,
		err := fetchSnapshotSizesKrews()
	if err != nil {
		pterm.Warning.Printf("Failed to fetch Krews data (mode=%s): %v\n", mode, err)
		results = append(results, providerSnapshotInfo{
			ProviderName: "Krews",
			Mode:         mode,
			TotalSize:    "unknown",
			BlockHeight:  "N/A",
			TimeAgo:      "N/A",
		})
	} else {
		var size, blockH, timeAgo string
		if mode == "pruned" {
			size = kPrunedSize
			blockH = kPrunedBlock
			timeAgo = kPrunedTimeAgo
		} else {
			size = kArchiveSize
			blockH = kArchiveBlock
			timeAgo = kArchiveTimeAgo
		}
		results = append(results, providerSnapshotInfo{
			ProviderName: "Krews",
			Mode:         mode,
			TotalSize:    size,
			BlockHeight:  blockH,
			TimeAgo:      timeAgo,
		})
	}

	// JNODE
	jPrunedSize, jArchiveSize,
		jPrunedBlock, jArchiveBlock,
		jPrunedTimeAgo, jArchiveTimeAgo,
		err := fetchSnapshotSizesJnode()
	if err != nil {
		pterm.Warning.Printf("Failed to fetch Jnode data (mode=%s): %v\n", mode, err)
		results = append(results, providerSnapshotInfo{
			ProviderName: "Jnode",
			Mode:         mode,
			TotalSize:    "unknown",
			BlockHeight:  "N/A",
			TimeAgo:      "N/A",
		})
	} else {
		var size, blockH, timeAgo string
		if mode == "pruned" {
			size = jPrunedSize
			blockH = jPrunedBlock
			timeAgo = jPrunedTimeAgo
		} else {
			size = jArchiveSize
			blockH = jArchiveBlock
			timeAgo = jArchiveTimeAgo
		}
		results = append(results, providerSnapshotInfo{
			ProviderName: "Jnode",
			Mode:         mode,
			TotalSize:    size,
			BlockHeight:  blockH,
			TimeAgo:      timeAgo,
		})
	}

	return results, nil
}

// -----------------------------------------------------------------------------
// Provider Selection Prompt
// -----------------------------------------------------------------------------

func selectSnapshotProvider(providersData []providerSnapshotInfo) (string, error) {
	if len(providersData) == 0 {
		return "", errors.New("no providers data found")
	}

	type providerDisplay struct {
		Name         string
		DisplayExtra string
		original     providerSnapshotInfo
	}

	var items []providerDisplay
	for _, pd := range providersData {
		// e.g. "( mode: pruned | size: 46.50G | height: 1555254 | 1h 17m ago )"
		extra := fmt.Sprintf("( mode: %s | size: %s | height: %s | %s )",
			pd.Mode,
			pd.TotalSize,
			pd.BlockHeight,
			pd.TimeAgo,
		)
		items = append(items, providerDisplay{
			Name:         pd.ProviderName,
			DisplayExtra: extra,
			original:     pd,
		})
	}

	templates := &promptui.SelectTemplates{
		Label:    "Select the snapshot provider",
		Active:   "✔ {{ .Name | cyan }} {{ .DisplayExtra | faint }}",
		Inactive: "  {{ .Name | cyan }} {{ .DisplayExtra | faint }}",
		Selected: "{{ .Name | green }} - {{ .DisplayExtra | faint }}",
	}

	prompt := promptui.Select{
		Label:     "Select the snapshot provider",
		Items:     items,
		Templates: templates,
		Size:      len(items),
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return items[i].Name, nil
}

// -----------------------------------------------------------------------------
// Time / Utils
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// ITROCKET
// -----------------------------------------------------------------------------

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

// fetchItrocketForMode fetches Itrocket data based on pruning mode
func fetchItrocketForMode(mode string) (string, string, string, error) {
	var urls []string
	if mode == "pruned" {
		urls = endpoints.Itrocket.Pruned
	} else {
		urls = endpoints.Itrocket.Archive
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

// -----------------------------------------------------------------------------
// KREWS
// -----------------------------------------------------------------------------

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

func fetchSnapshotSizesKrews() (
	prunedSize, archiveSize string,
	prunedBlockHeight, archiveBlockHeight string,
	prunedTimeAgo, archiveTimeAgo string,
	err error,
) {
	resp, err := http.Get(endpoints.Krews)
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

// -----------------------------------------------------------------------------
// Download Steps (Itrocket, Krews, etc.)
// -----------------------------------------------------------------------------

func downloadSnapshotItrocket(homeDir, mode string) error {
	var serverURL string
	if strings.ToLower(mode) == "pruned" {
		serverURL = bestPrunedServerURL
	} else {
		serverURL = bestArchiveServerURL
	}
	if serverURL == "" {
		return errors.New("no best server URL found for the selected mode")
	}

	printInfo(fmt.Sprintf("Fetching snapshot names from Itrocket (%s)...", serverURL))
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

	printInfo("Installing required packages for Itrocket snapshot...")
	if err := runCommand("sudo apt install curl tmux jq lz4 unzip -y"); err != nil {
		return err
	}

	printInfo("Stopping Story and Story-Geth services...")
	if err := runCommand("sudo systemctl stop story story-geth"); err != nil {
		return err
	}

	printInfo("Backup priv_validator_state.json...")
	if err := runCommand(fmt.Sprintf("cp %s/.story/story/data/priv_validator_state.json %s/.story/story/priv_validator_state.json.backup", homeDir, homeDir)); err != nil {
		return err
	}

	printInfo("Removing old Story data...")
	storySnapshotURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(serverURL, "/.current_state.json"), snapshotState.SnapshotName)
	storySnapshotPath := filepath.Join(homeDir, ".story", "story_snapshot.tar.lz4")

	printInfo("Downloading Story snapshot...")
	if err := downloadFileWithProgress(storySnapshotURL, storySnapshotPath); err != nil {
		return err
	}

	printInfo("Extracting Story snapshot...")
	if err := decompressAndExtractLz4Tar(storySnapshotPath, filepath.Join(homeDir, ".story", "story")); err != nil {
		return err
	}
	if err := os.Remove(storySnapshotPath); err != nil {
		return err
	}

	printInfo("Removing old Geth data and downloading new Geth snapshot...")
	gethSnapshotURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(serverURL, "/.current_state.json"), snapshotState.SnapshotGethName)
	gethSnapshotPath := filepath.Join(homeDir, ".story", "geth_snapshot.tar.lz4")

	printInfo("Downloading Geth snapshot...")
	if err := downloadFileWithProgress(gethSnapshotURL, gethSnapshotPath); err != nil {
		return err
	}
	printInfo("Extracting Geth snapshot...")
	if err := decompressAndExtractLz4Tar(gethSnapshotPath, filepath.Join(homeDir, ".story", "geth", "iliad", "geth")); err != nil {
		return err
	}
	if err := os.Remove(gethSnapshotPath); err != nil {
		return err
	}

	printInfo("Restoring priv_validator_state.json...")
	if err := runCommand(fmt.Sprintf("mv %s/.story/story/priv_validator_state.json.backup %s/.story/story/data/priv_validator_state.json", homeDir, homeDir)); err != nil {
		return err
	}

	printInfo("Starting Story and Story-Geth services...")
	if err := runCommand("sudo systemctl restart story story-geth"); err != nil {
		return err
	}

	printSuccess("Snapshot successfully downloaded and applied from Itrocket.")
	return nil
}

func downloadSnapshotKrews(homeDir, pruningMode string) error {
	snapshotName := fmt.Sprintf("story_testnet_%s_snapshot", pruningMode)
	snapshotURL := fmt.Sprintf("krews-snapshot:krews-1-eu/%s", snapshotName)
	destDir := filepath.Join(homeDir, ".story")

	printInfo("Installing and configuring Rclone for Krews snapshot...")
	err := installAndConfigureRcloneKrews(homeDir)
	if err != nil {
		return err
	}

	printInfo("Backup priv_validator_state.json...")
	err = runCommand(fmt.Sprintf("cp %s/.story/story/data/priv_validator_state.json %s/.story/story/priv_validator_state.json.backup ", homeDir, homeDir))
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

	printInfo("Restoring priv_validator_state.json...")
	err = runCommand(fmt.Sprintf("mv %s/.story/story/priv_validator_state.json.backup %s/.story/story/data/priv_validator_state.json", homeDir, homeDir))
	if err != nil {
		return err
	}

	printSuccess("Snapshot successfully downloaded from Krews.")
	return nil
}

// -----------------------------------------------------------------------------
// Download Snapshots to a Path (Used when --output-path is provided)
// -----------------------------------------------------------------------------

func downloadSnapshotToPathItrocket(mode, path string) error {
	var urls []string
	if mode == "pruned" {
		urls = endpoints.Itrocket.Pruned
	} else {
		urls = endpoints.Itrocket.Archive
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

	printInfo(fmt.Sprintf("Downloading Itrocket Story snapshot from %s to %s...", storySnapshotURL, storyDestPath))
	err = downloadFileWithProgress(storySnapshotURL, storyDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Itrocket Story snapshot: %v", err)
	}

	printInfo(fmt.Sprintf("Downloading Itrocket Geth snapshot from %s to %s...", gethSnapshotURL, gethDestPath))
	err = downloadFileWithProgress(gethSnapshotURL, gethDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Itrocket Geth snapshot: %v", err)
	}

	return nil
}

func downloadSnapshotToPathKrews(mode, path string) error {
	snapshotName := fmt.Sprintf("story_testnet_%s_snapshot", mode)
	snapshotURL := fmt.Sprintf("krews-snapshot:krews-1-eu/%s", snapshotName)

	cmd := exec.Command("rclone", "copy", "--no-check-certificate", "--transfers=6", "--checkers=6", snapshotURL, path+"/"+snapshotName, "--progress")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("rclone copy failed: %v", err)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Jnode Structures & Download Functions
// -----------------------------------------------------------------------------

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

// fetchSnapshotSizesJnode fetches snapshot sizes and details from jnode API
func fetchSnapshotSizesJnode() (
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

// downloadSnapshotToPathJnode downloads the Jnode snapshot to a specified path without applying it
func downloadSnapshotToPathJnode(mode, path string) error {
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

	printInfo(fmt.Sprintf("Downloading Jnode Story snapshot from %s to %s...", storySnapshotURL, storyDestPath))
	err = downloadFileWithProgress(storySnapshotURL, storyDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Jnode Story snapshot: %v", err)
	}

	printInfo(fmt.Sprintf("Downloading Jnode Geth snapshot from %s to %s...", gethSnapshotURL, gethDestPath))
	err = downloadFileWithProgress(gethSnapshotURL, gethDestPath)
	if err != nil {
		return fmt.Errorf("failed to download Jnode Geth snapshot: %v", err)
	}

	return nil
}

// downloadSnapshotJnode downloads and applies the Jnode snapshot
func downloadSnapshotJnode(homeDir, mode string) error {
	printInfo("Installing required packages for Jnode snapshot...")
	if err := runCommand("sudo apt-get install wget lz4 aria2 pv -y"); err != nil {
		return err
	}

	printInfo("Stopping Story and Story-Geth services...")
	if err := runCommand("sudo systemctl stop story story-geth"); err != nil {
		return err
	}

	printInfo("Backing up priv_validator_state.json...")
	if err := runCommand(fmt.Sprintf("cp %s/.story/story/data/priv_validator_state.json %s/.story/priv_validator_state.json.backup", homeDir, homeDir)); err != nil {
		return err
	}

	printInfo("Removing old Story and Geth data...")
	if err := runCommand(fmt.Sprintf("rm -f %s/.story/story/data/*", homeDir)); err != nil {
		return err
	}
	if err := runCommand(fmt.Sprintf("rm -rf %s/.story/geth/odyssey/geth/chaindata", homeDir)); err != nil {
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

	printInfo("Downloading Story snapshot...")
	storySnapshotPath := filepath.Join(homeDir, "Story_snapshot.lz4")
	if err := runCommand(fmt.Sprintf("aria2c -x 16 -s 16 -k 1M %s -o %s", storySnapshotURL, storySnapshotPath)); err != nil {
		return fmt.Errorf("failed to download Story snapshot: %v", err)
	}

	printInfo("Downloading Geth snapshot...")
	gethSnapshotPath := filepath.Join(homeDir, "Geth_snapshot.lz4")
	if err := runCommand(fmt.Sprintf("aria2c -x 16 -s 16 -k 1M %s -o %s", gethSnapshotURL, gethSnapshotPath)); err != nil {
		return fmt.Errorf("failed to download Geth snapshot: %v", err)
	}

	printInfo("Extracting Story snapshot...")
	if err := runCommand(fmt.Sprintf("lz4 -d -c %s | pv | sudo tar xv -C %s/.story/story/ > /dev/null", storySnapshotPath, homeDir)); err != nil {
		return fmt.Errorf("failed to extract Story snapshot: %v", err)
	}
	if err := runCommand(fmt.Sprintf("rm -f %s", storySnapshotPath)); err != nil {
		return err
	}

	printInfo("Extracting Geth snapshot...")
	if err := runCommand(fmt.Sprintf("lz4 -d -c %s | pv | sudo tar xv -C %s/.story/geth/odyssey/geth/ > /dev/null", gethSnapshotPath, homeDir)); err != nil {
		return fmt.Errorf("failed to extract Geth snapshot: %v", err)
	}
	if err := runCommand(fmt.Sprintf("rm -f %s", gethSnapshotPath)); err != nil {
		return err
	}

	printInfo("Restoring priv_validator_state.json...")
	if err := runCommand(fmt.Sprintf("cp %s/.story/priv_validator_state.json.backup %s/.story/story/data/priv_validator_state.json", homeDir, homeDir)); err != nil {
		return err
	}

	printInfo("Starting Story and Story-Geth services...")
	if err := runCommand("sudo systemctl restart story story-geth"); err != nil {
		return err
	}

	printSuccess("Snapshot successfully downloaded and applied from Jnode.")
	return nil
}
