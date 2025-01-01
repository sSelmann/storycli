package snapshot

import (
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// providersCmd represents the snapshot providers subcommand
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available snapshot providers and their data",
	Long:  `List available snapshot providers and display their snapshot data in a table format.`,
	RunE:  runListProviders,
}

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
