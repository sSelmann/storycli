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

	"github.com/sSelmann/storycli/internal/config" // <-- Config paketini içe aktarın
)

// snapshotCmd represents the main snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots for Story node",
	Long:  `Manage snapshots for Story node with subcommands like download and providers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Eğer snapshot komutu kendi başına çağrılırsa, kullanıcıyı yönlendirin
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

// Global vars
var (
	homeDirFlag          string
	pruningMode          string
	bestPrunedServerURL  string
	bestArchiveServerURL string

	// Seçilen provider (prompt'tan)
	selectedProvider string

	// API endpoint'lerini tutacak değişken
	endpoints config.Endpoints
)

func init() {
	// Ana snapshot komutuna subcommand'ları ekleyin
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(downloadCmd)
	snapshotCmd.AddCommand(providersCmd)

	// download subcommand için flag ekleyebilirsiniz (örneğin, home directory)
	downloadCmd.Flags().StringVar(&homeDirFlag, "home", defaultHomeDir(), "Home directory for Story node")

	downloadCmd.Flags().String("output-path", "", "Download snapshot directly to the specified path without setup it")

	// Varsayılan endpoint'leri atayın
	endpoints = config.DefaultEndpoints()
}

func defaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		os.Exit(1)
	}
	return filepath.Join(home)
}

// runDownloadSnapshot fonksiyonu mevcut snapshot komutunun RunE'si olacak
func runDownloadSnapshot(cmd *cobra.Command, args []string) error {
	// Flag değerini al
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

	// 2) Fetch data for all providers, but specifically for the chosen pruning mode
	printInfo(fmt.Sprintf("Fetching snapshot data for providers (mode=%s)...", pruningMode))
	providersData, err := fetchAllProvidersDataForMode(pruningMode)
	if err != nil {
		printWarning(fmt.Sprintf("Failed to fetch providers data: %v", err))
		// Yine de devam edelim, eğer veriler varsa
	}

	if len(providersData) == 0 {
		return errors.New("no snapshot data found for any provider")
	}

	// 3) Show provider selection prompt (with the relevant data for chosen pruning mode)
	selectedProvider, err = selectSnapshotProvider(providersData)
	if err != nil {
		return err
	}

	// Eğer output-path flag'ı set edilmişse, direkt olarak indir ve işlemleri atla
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

	// 5) Download the snapshot from the chosen provider & chosen pruning mode
	switch selectedProvider {
	case "Itrocket":
		if err := downloadSnapshotItrocket(homeDirFlag, pruningMode); err != nil {
			return err
		}
	case "Krews":
		if err := downloadSnapshotKrews(homeDirFlag, pruningMode); err != nil {
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

// runListProviders fonksiyonu providers subcommand'unun RunE'si olacak
// runListProviders fonksiyonu providers subcommand'unun RunE'si olacak
func runListProviders(cmd *cobra.Command, args []string) error {
	// Pruning modlarını sabit olarak tanımlayın
	modes := []string{"pruned", "archive"}

	// Sağlayıcı verilerini çekin için tüm modları kullanın
	providersData, err := fetchAllProvidersDataForModes(modes)
	if err != nil {
		return fmt.Errorf("failed to fetch providers data: %v", err)
	}

	if len(providersData) == 0 {
		return errors.New("no providers data available")
	}

	// Verileri modlara göre ayırın
	prunedProviders := []providerSnapshotInfo{}
	archiveProviders := []providerSnapshotInfo{}

	for _, pd := range providersData {
		if pd.Mode == "pruned" {
			prunedProviders = append(prunedProviders, pd)
		} else if pd.Mode == "archive" {
			archiveProviders = append(archiveProviders, pd)
		}
	}

	// Pruned Providers Tablosunu Oluşturun
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

	// Archive Providers Tablosunu Oluşturun
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

// ---------------------------------------------------
// FETCH DATA FOR MULTIPLE PRUNING MODES
// ---------------------------------------------------

// providerSnapshotInfo : info to show in provider prompt
type providerSnapshotInfo struct {
	ProviderName string
	Mode         string
	TotalSize    string // sum of snapshot_size + geth_snapshot_size (for Itrocket)
	BlockHeight  string
	TimeAgo      string
}

// fetchAllProvidersDataForModes collects data for the specified pruning modes (pruned and archive)
func fetchAllProvidersDataForModes(modes []string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	for _, mode := range modes {
		// -- ITROCKET --
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

		// -- KREWS --
		kPrunedSize, kArchiveSize,
			kPrunedBlock, kArchiveBlock,
			kPrunedTimeAgo, kArchiveTimeAgo,
			err := fetchSnapshotSizesKrews()
		if err != nil {
			pterm.Warning.Printf("Failed to fetch Krews data (mode=%s): %v\n", mode, err)
			// Even on error, add a "Krews" item
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
	}

	return results, nil
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

	// Assign the best server URL based on pruning mode
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

// sumSnapshotSizes parses e.g. "8.2G" + "44G" => "52.2G"
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
	// We assume something like "8.2G" or "44G". We remove the "G"
	if sizeStr == "" {
		return -1
	}
	s := strings.ToUpper(sizeStr)
	if strings.HasSuffix(s, "G") {
		s = strings.TrimSuffix(s, "G")
	} else if strings.HasSuffix(s, "GB") {
		s = strings.TrimSuffix(s, "GB")
	} else {
		// unsupported format
		return -1
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return val
}

// ---------------------------------------------------
// PROVIDER SELECTION PROMPT
// ---------------------------------------------------

func selectSnapshotProvider(providersData []providerSnapshotInfo) (string, error) {
	if len(providersData) == 0 {
		return "", errors.New("no providers data found")
	}

	// "faint" yazı oluşturmak için, providerDisplay adında bir struct kullanıyoruz
	type providerDisplay struct {
		Name         string
		DisplayExtra string
		original     providerSnapshotInfo
	}

	// Hem name hem de faint extra'yı hazırlıyoruz
	items := make([]providerDisplay, 0, len(providersData))
	for _, pd := range providersData {
		// Örneğin "( mode: pruned | size: 46.50G | height: 1555254 | 1h 17m ago )" benzeri
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
		// Prompt'un en üstünde gözükecek sabit metin
		Label: "Select the snapshot provider",

		// Kullanıcı o satırın üstündeyken gösterilen format
		Active: "✔ {{ .Name | cyan }} {{ .DisplayExtra | faint }}",
		// Seçili değilken gösterilen format
		Inactive: "  {{ .Name | cyan }} {{ .DisplayExtra | faint }}",
		// Kullanıcı Enter’a bastığında seçilen öğenin nasıl gösterileceği
		Selected: "{{ .Name | green }} - {{ .DisplayExtra | faint }}",
	}

	prompt := promptui.Select{
		Label:     "Select the snapshot provider",
		Items:     items,
		Templates: templates,

		// Listede kaç öğe gösterileceği
		Size: len(items),
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return items[i].Name, nil
}

// ---------------------------------------------------
// TIME / UTILS
// ---------------------------------------------------

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

// ---------------------------------------------------
// ITROCKET
// ---------------------------------------------------

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

// ---------------------------------------------------
// KREWS
// ---------------------------------------------------

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
		// 1) boyut
		sz := snapshot.Size
		// 2) block height
		blk := snapshot.Block.String()
		// 3) "snapshot_date"
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

func parseKrewsSnapshotDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}
	// "26 Dec 2024, 18:17:50" => layout: "02 Jan 2006, 15:04:05"
	const krewsLayout = "02 Jan 2006, 15:04:05"

	t, err := time.Parse(krewsLayout, dateStr)
	if err != nil {
		// parse edilemezse "N/A"
		pterm.Warning.Printf("Could not parse Krews snapshot_date=%s: %v\n", dateStr, err)
		return "N/A"
	}
	// Time farkını hesapla
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

// ---------------------------------------------------
// DOWNLOAD STEPS (ITROCKET, KREWS, ETC.)
// ---------------------------------------------------

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

	// Example steps for installing packages, removing old data, etc.
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
	// Existing Krews snapshot process
	snapshotName := fmt.Sprintf("story_testnet_%s_snapshot", pruningMode)
	snapshotURL := fmt.Sprintf("krews-snapshot:krews-1-eu/%s", snapshotName)
	destDir := filepath.Join(homeDir, ".story")

	// Install and configure Rclone specifically for Krews
	printInfo("Installing and configuring Rclone for Krews snapshot...")
	err := installAndConfigureRcloneKrews(homeDir)
	if err != nil {
		return err
	}

	// Backup priv_validator_state.json
	printInfo("Backup priv_validator_state.json...")
	err = runCommand(fmt.Sprintf("cp %s/.story/story/data/priv_validator_state.json %s/.story/story/priv_validator_state.json.backup ", homeDir, homeDir))
	if err != nil {
		return err
	}

	// Run rclone with --progress, output visible
	cmd := exec.Command("rclone", "copy", "--no-check-certificate", "--transfers=6", "--checkers=6", snapshotURL, destDir, "--progress")

	// Connect rclone's stdout and stderr to the program's stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	// Restore priv_validator_state.json
	printInfo("Restoring priv_validator_state.json...")
	err = runCommand(fmt.Sprintf("mv %s/.story/story/priv_validator_state.json.backup %s/.story/story/data/priv_validator_state.json", homeDir, homeDir))
	if err != nil {
		return err
	}

	printSuccess("Snapshot successfully downloaded from Krews.")
	return nil
}

// fetchAllProvidersDataForMode collects data for the *chosen pruning mode* (either pruned or archive)
func fetchAllProvidersDataForMode(mode string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	// -- ITROCKET --
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

	// -- KREWS --
	kPrunedSize, kArchiveSize,
		kPrunedBlock, kArchiveBlock,
		kPrunedTimeAgo, kArchiveTimeAgo,
		err := fetchSnapshotSizesKrews()
	if err != nil {
		pterm.Warning.Printf("Failed to fetch Krews data (mode=%s): %v\n", mode, err)
		// Even on error, add a "Krews" item
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

	return results, nil
}

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

	// Determine the snapshot URL
	var snapshotURL string
	if mode == "pruned" {
		snapshotURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(best.serverURL, "/.current_state.json"), best.state.SnapshotName)
	} else {
		snapshotURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(best.serverURL, "/.current_state.json"), best.state.SnapshotName)
	}

	// Download the snapshot to the specified path
	err = downloadFileWithProgress(snapshotURL, path+"/"+best.state.SnapshotName)
	if err != nil {
		return fmt.Errorf("failed to download snapshot from Itrocket: %v", err)
	}

	return nil
}

func downloadSnapshotToPathKrews(mode, path string) error {
	// Determine the snapshot name based on mode
	snapshotName := fmt.Sprintf("story_testnet_%s_snapshot", mode)
	// Determine the snapshot URL for Krews
	snapshotURL := fmt.Sprintf("krews-snapshot:krews-1-eu/%s", snapshotName)

	// Download the snapshot to the specified path using Rclone
	// Assuming Rclone is installed and configured correctly for Krews
	cmd := exec.Command("rclone", "copy", "--no-check-certificate", "--transfers=6", "--checkers=6", snapshotURL, path+"/"+snapshotName, "--progress")

	// Connect rclone's stdout and stderr to the program's stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("rclone copy failed: %v", err)
	}

	return nil
}
