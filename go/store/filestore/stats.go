package filestore

import (
	"fmt"

	"github.com/romayengineer/facebook_marketplace_extractor/go/core/jsonutil"
)

// DescriptionStats summarizes how many detail records have a description.
// A listing counts as empty only when Description is JSON null; an empty string
// counts as present.
type DescriptionStats struct {
	Total           int
	Empty           int
	WithDescription int
}

func (s DescriptionStats) EmptyPercent() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Empty) / float64(s.Total) * 100
}

func (s DescriptionStats) WithDescriptionPercent() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.WithDescription) / float64(s.Total) * 100
}

// ComputeDescriptionStats walks data/__Details and counts records by whether their
// Description is null.
func ComputeDescriptionStats() DescriptionStats {
	stats := DescriptionStats{}
	ForEachDetail(func(_ string, jsonData map[string]any) bool {
		stats.Total++
		if jsonutil.GetKey(jsonData, "Description") == nil {
			stats.Empty++
		}
		return true
	}, false)
	stats.WithDescription = stats.Total - stats.Empty
	return stats
}

// PrintDescriptionStats writes a human-readable description report to stdout.
func PrintDescriptionStats(stats DescriptionStats) {
	fmt.Println("data/__Details description report")
	fmt.Printf("  total listings:    %d\n", stats.Total)
	fmt.Printf("  empty (null):      %d (%.2f%%)\n", stats.Empty, stats.EmptyPercent())
	fmt.Printf("  with description:  %d (%.2f%%)\n", stats.WithDescription, stats.WithDescriptionPercent())
}
