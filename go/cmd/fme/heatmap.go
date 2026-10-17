package main

import (
	"fmt"

	"github.com/romayengineer/fme/go/analytics/pipeline"
	"github.com/romayengineer/fme/go/core/logging"
)

// runHeatmap aggregates product locations from data/__Details into a grid and
// writes heatmap.geojson plus a self-contained heatmap.html. See
// `-action=heatmap -cell=<deg> -out=<base> -dir=<details dir>`.
func runHeatmap(flags Flags) {
	result, err := pipeline.RunFileHeatmap(pipeline.Options{
		Dir:     flags.dir,
		CellDeg: flags.cellDeg,
		OutBase: flags.out,
	})
	if err != nil {
		logging.LogError0("runHeatmap", "heat map failed", "error", err)
		return
	}

	fmt.Printf("heat map written\n")
	fmt.Printf("  geoJSON:   %s\n", result.GeoJSONPath)
	fmt.Printf("  html:      %s\n", result.HTMLPath)
	fmt.Printf("  products:  %d (of %d scanned, %d skipped: no/!valid coords)\n",
		result.Stats.Valid, result.Stats.Scanned, result.Stats.Skipped)
	fmt.Printf("  cells:     %d (peak %d, cell %.4f deg)\n",
		len(result.Grid.Cells), result.Grid.MaxCount, result.Grid.Options.CellDeg)

	logging.LogInfo0("runHeatmap", "heat map written",
		"geojson", result.GeoJSONPath, "html", result.HTMLPath,
		"products", result.Stats.Valid, "cells", len(result.Grid.Cells))
}
