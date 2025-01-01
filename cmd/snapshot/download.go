package snapshot

import (
	"errors"
	"fmt"

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
