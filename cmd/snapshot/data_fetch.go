package snapshot

import (
	"github.com/pterm/pterm"
	"github.com/sSelmann/storycli/snapshot_providers/itrocket"
	"github.com/sSelmann/storycli/snapshot_providers/jnode"
	"github.com/sSelmann/storycli/snapshot_providers/krews"
	"github.com/sSelmann/storycli/snapshot_providers/mandragora"
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
		totalSizeIt, blockHeightIt, timeAgoIt, err := itrocket.FetchItrocketForMode(mode, endpoints.Itrocket)
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

		// MANDRAGORA
		mPrunedSize, mArchiveSize,
			mPrunedBlock, mArchiveBlock,
			mPrunedTimeAgo, mArchiveTimeAgo,
			err := mandragora.FetchSnapshotSizesMandragora()
		if err != nil {
			pterm.Warning.Printf("Failed to fetch Mandragora data (mode=%s): %v\n", mode, err)
			results = append(results, providerSnapshotInfo{
				ProviderName: "Mandragora",
				Mode:         mode,
				TotalSize:    "unknown",
				BlockHeight:  "N/A",
				TimeAgo:      "N/A",
			})
		} else {
			var size, blockH, timeAgo string
			if mode == "pruned" {
				size = mPrunedSize
				blockH = mPrunedBlock
				timeAgo = mPrunedTimeAgo
			} else {
				size = mArchiveSize
				blockH = mArchiveBlock
				timeAgo = mArchiveTimeAgo
			}
			results = append(results, providerSnapshotInfo{
				ProviderName: "Mandragora",
				Mode:         mode,
				TotalSize:    size,
				BlockHeight:  blockH,
				TimeAgo:      timeAgo,
			})
		}

		// KREWS
		kPrunedSize, kArchiveSize,
			kPrunedBlock, kArchiveBlock,
			kPrunedTimeAgo, kArchiveTimeAgo,
			err := krews.FetchSnapshotSizesKrews(endpoints.Krews)
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
	}

	return results, nil
}

func fetchAllProvidersDataForMode(mode string) ([]providerSnapshotInfo, error) {
	var results []providerSnapshotInfo

	// ITROCKET
	totalSizeIt, blockHeightIt, timeAgoIt, err := itrocket.FetchItrocketForMode(mode, endpoints.Itrocket)
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

	// MANDRAGORA
	mPrunedSize, mArchiveSize,
		mPrunedBlock, mArchiveBlock,
		mPrunedTimeAgo, mArchiveTimeAgo,
		err := mandragora.FetchSnapshotSizesMandragora()
	if err != nil {
		pterm.Warning.Printf("Failed to fetch Mandragora data (mode=%s): %v\n", mode, err)
		results = append(results, providerSnapshotInfo{
			ProviderName: "Mandragora",
			Mode:         mode,
			TotalSize:    "unknown",
			BlockHeight:  "N/A",
			TimeAgo:      "N/A",
		})
	} else {
		var size, blockH, timeAgo string
		if mode == "pruned" {
			size = mPrunedSize
			blockH = mPrunedBlock
			timeAgo = mPrunedTimeAgo
		} else {
			size = mArchiveSize
			blockH = mArchiveBlock
			timeAgo = mArchiveTimeAgo
		}
		results = append(results, providerSnapshotInfo{
			ProviderName: "Mandragora",
			Mode:         mode,
			TotalSize:    size,
			BlockHeight:  blockH,
			TimeAgo:      timeAgo,
		})
	}

	// KREWS
	kPrunedSize, kArchiveSize,
		kPrunedBlock, kArchiveBlock,
		kPrunedTimeAgo, kArchiveTimeAgo,
		err := krews.FetchSnapshotSizesKrews(endpoints.Krews)
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

	return results, nil
}
