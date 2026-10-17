package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/romayengineer/fme/go/analytics/pipeline"
	"github.com/romayengineer/fme/go/core/logging"
)

// runHeatmapServe builds the heat map and serves the output directory over HTTP.
//
// Opening heatmap.html directly as a file:// URL often makes the browser send no
// (or a null) origin, and tile servers such as OpenStreetMap reject those
// requests, so the basemap shows "Access blocked". Serving the same page from
// 127.0.0.1 gives the browser a real origin and the tiles load normally.
func runHeatmapServe(flags Flags) {
	result, err := pipeline.RunFileHeatmap(pipeline.Options{
		Dir:     flags.dir,
		CellDeg: flags.cellDeg,
		OutBase: flags.out,
	})
	if err != nil {
		logging.LogError0("runHeatmapServe", "heat map failed", "error", err)
		return
	}

	fmt.Printf("heat map written\n")
	fmt.Printf("  products:  %d (of %d scanned, %d skipped)\n",
		result.Stats.Valid, result.Stats.Scanned, result.Stats.Skipped)
	fmt.Printf("  cells:     %d (peak %d, cell %.4f deg)\n",
		len(result.Grid.Cells), result.Grid.MaxCount, result.Grid.Options.CellDeg)

	dir := filepath.Dir(result.HTMLPath)
	url := fmt.Sprintf("http://%s/%s", flags.addr, filepath.Base(result.HTMLPath))
	fmt.Printf("\nserving %s at %s\n", dir, url)
	fmt.Printf("press Ctrl+C to stop\n")

	logging.LogInfo0("runHeatmapServe", "serving heat map", "dir", dir, "url", url)

	if err := http.ListenAndServe(flags.addr, noCache(http.FileServer(http.Dir(dir)))); err != nil {
		logging.LogError0("runHeatmapServe", "http server stopped", "error", err)
	}
}

// noCache stops the browser from serving a stale heat map after a re-run.
func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
