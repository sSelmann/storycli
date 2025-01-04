package snapshot

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/snapshot_providers/itrocket"
	"github.com/sSelmann/storycli/snapshot_providers/jnode"
	"github.com/sSelmann/storycli/snapshot_providers/krews"
	"github.com/sSelmann/storycli/utils/bash"
	"github.com/spf13/cobra"
)

// downloadCmd represents the snapshot download subcommand
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download and apply snapshots",
	Long:  `Download and apply snapshots from selected providers to initialize or update your Story node.`,
	RunE:  runDownloadSnapshot,
}

func runDownloadSnapshot(cmd *cobra.Command, args []string) error {
	outputPath, err := cmd.Flags().GetString("output-path")
	if err != nil {
		return fmt.Errorf("failed to parse output-path: %w", err)
	}
	return RunDownloadSnapshotCore(pruningMode, outputPath, false, homeDirFlag)
}

func CallRunDownloadSnapshotManually(pruningMode string, homedir string) error {
	return RunDownloadSnapshotCore(pruningMode, "", true, homedir+".story")
}

func RunDownloadSnapshotCore(pruningMode, outputPath string, isManual bool, storyDir string) error {
	if !isManual {
		PruningModeInformation()
	}

	if pruningMode == "" {
		var err error
		pruningMode, err = SelectPruningMode()
		if err != nil {
			return fmt.Errorf("failed to select pruning mode: %w", err)
		}
	}

	pterm.Info.Println(fmt.Sprintf("Fetching snapshot data for providers (mode=%s)...", pruningMode))
	providersData, err := fetchAllProvidersDataForMode(pruningMode)
	if err != nil || len(providersData) == 0 {
		return fmt.Errorf("no snapshot data found: %w", err)
	}

	selectedProvider, err := selectSnapshotProvider(providersData)
	if err != nil {
		return err
	}

	if outputPath != "" {
		pterm.Info.Println(fmt.Sprintf("Downloading snapshot from %s to %s...", selectedProvider, outputPath))
		return downloadToPath(selectedProvider, pruningMode, outputPath)
	} else if !isManual {
		storyDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		storyDir = storyDir + ".story"
		sDir := storyDir + "/story"
		gDir := storyDir + "/geth"
		if _, err := os.Stat(sDir); err == nil {
			pterm.Error.Println("Story path not found: " + sDir)
			return err
		}
		if _, err := os.Stat(gDir); err == nil {
			pterm.Error.Println("Geth path not found:" + gDir)
			return err
		}
	}

	if !isManual {
		pterm.Info.Println("Stopping services...")
		if err := bash.RunCommand("sudo", "systemctl", "stop", "story", "story-geth"); err != nil {
			return err
		}
	}

	return downloadAndApplySnapshot(selectedProvider, pruningMode)
}

func downloadToPath(provider, mode, path string) error {
	switch provider {
	case "Itrocket":
		return itrocket.DownloadSnapshotToPathItrocket(mode, path, endpoints.Itrocket)
	case "Krews":
		return krews.DownloadSnapshotToPathKrews(mode, path)
	case "Jnode":
		return jnode.DownloadSnapshotToPathJnode(mode, path, endpoints.Jnode)
	default:
		return errors.New("unsupported provider")
	}
}

func downloadAndApplySnapshot(provider, mode string) error {
	switch provider {
	case "Itrocket":
		return itrocket.DownloadSnapshotItrocket(homeDirFlag, mode)
	case "Krews":
		return krews.DownloadSnapshotKrews(homeDirFlag, mode)
	case "Jnode":
		return jnode.DownloadSnapshotJnode(homeDirFlag, mode, endpoints.Jnode)
	default:
		return errors.New("unsupported provider")
	}
}

func PruningModeInformation() {
	pterm.Info.Println("Pruning Mode Information:")
	pterm.Info.Println(" - Pruned Mode: Stores only recent blocks, reducing disk usage.")
	pterm.Info.Println(" - Archive Mode: Stores the entire chain history, requiring more disk space.")
}

func SelectPruningMode() (string, error) {
	if pruningMode == "" {
		pmPrompt := promptui.Select{
			Label: "Select the pruning mode",
			Items: []string{"pruned", "archive"},
		}
		_, pmResult, err := pmPrompt.Run()
		if err != nil {
			return "", err
		}
		pruningMode = pmResult
	}
	return pruningMode, nil
}

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
