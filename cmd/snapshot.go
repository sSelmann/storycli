package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/sSelmann/storycli/snapshot_providers/itrocket"
	"github.com/sSelmann/storycli/snapshot_providers/jnode"
	"github.com/sSelmann/storycli/snapshot_providers/krews"
	"github.com/sSelmann/storycli/utils/bash"
)

// -----------------------------------------------------------------------------
// Global Variables
// -----------------------------------------------------------------------------

var (
	// homeDirFlag is used to store the home directory path (passed via --home).
	homeDirFlag string

	// pruningMode represents the selected pruning mode ("pruned" or "archive").
	pruningMode string

	// selectedProvider stores the chosen provider (Itrocket, Krews, Jnode).
	selectedProvider string
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
	pterm.Info.Println("Pruning Mode Information:")
	pterm.Info.Println(" - Pruned Mode: Stores only recent blocks, reducing disk usage.")
	pterm.Info.Println(" - Archive Mode: Stores the entire chain history, requiring more disk space.")

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
	pterm.Info.Printf(fmt.Sprintf("Fetching snapshot data for providers (mode=%s)...", pruningMode))
	providersData, err := fetchAllProvidersDataForMode(pruningMode) // includes Jnode
	if err != nil {
		pterm.Warning.Printf(fmt.Sprintf("Failed to fetch providers data: %v", err))
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
		pterm.Info.Printf(fmt.Sprintf("Downloading snapshot from %s in %s mode to %s...", selectedProvider, pruningMode, outputPath))

		switch selectedProvider {
		case "Itrocket":
			err = itrocket.DownloadSnapshotToPathItrocket(pruningMode, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download snapshot from Itrocket: %v", err)
			}
		case "Krews":
			err = krews.DownloadSnapshotToPathKrews(pruningMode, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download snapshot from Krews: %v", err)
			}
		case "Jnode":
			err = jnode.DownloadSnapshotToPathJnode(pruningMode, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download snapshot from Jnode: %v", err)
			}
		default:
			return errors.New("provider not supported yet")
		}

		pterm.Success.Printf(fmt.Sprintf("Snapshot successfully downloaded to %s from %s.", outputPath, selectedProvider))
		return nil
	}

	// 4) Stop services
	pterm.Info.Printf("Stopping services...")
	if err := bash.RunCommand("sudo systemctl stop story story-geth"); err != nil {
		return err
	}

	// 5) Download the snapshot from chosen provider & mode
	switch selectedProvider {
	case "Itrocket":
		if err := itrocket.DownloadSnapshotItrocket(homeDirFlag, pruningMode); err != nil {
			return err
		}
	case "Krews":
		if err := krews.DownloadSnapshotKrews(homeDirFlag, pruningMode); err != nil {
			return err
		}
	case "Jnode":
		if err := jnode.DownloadSnapshotJnode(homeDirFlag, pruningMode); err != nil {
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
		totalSizeIt, blockHeightIt, timeAgoIt, err := itrocket.FetchItrocketForMode(mode)
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
			err := krews.FetchSnapshotSizesKrews()
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
			err := jnode.FetchSnapshotSizesJnode()
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
	totalSizeIt, blockHeightIt, timeAgoIt, err := itrocket.FetchItrocketForMode(mode)
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
		err := krews.FetchSnapshotSizesKrews()
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
		err := jnode.FetchSnapshotSizesJnode()
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
