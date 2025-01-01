package snapshot

import (
	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/snapshot_providers/itrocket"
	"github.com/sSelmann/storycli/snapshot_providers/jnode"
	"github.com/sSelmann/storycli/snapshot_providers/krews"
)

// providerSnapshotInfo holds data displayed for each provider
type providerSnapshotInfo struct {
	ProviderName string
	Mode         string
	TotalSize    string // sum of snapshot_size + geth_snapshot_size (for Itrocket)
	BlockHeight  string
	TimeAgo      string
}

func fetchAllProvidersDataForModes(modes []string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	for _, mode := range modes {
		// ITROCKET
		totalSizeIt, blockHeightIt, timeAgoIt, err := itrocket.FetchItrocketForMode(mode)
		if err != nil {
			pterm.Warning.Println("Failed to fetch Itrocket data (mode=%s): %v\n", mode, err)
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

func fetchAllProvidersDataForMode(mode string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	// ITROCKET
	totalSizeIt, blockHeightIt, timeAgoIt, err := itrocket.FetchItrocketForMode(mode)
	if err != nil {
		pterm.Warning.Println("Failed to fetch Itrocket data (mode=%s): %v\n", mode, err)
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
