// Package pipeline wires a PointSource to the heatmap aggregation and writers.
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/romayengineer/fme/go/analytics"
	"github.com/romayengineer/fme/go/analytics/heatmap"
	"github.com/romayengineer/fme/go/analytics/source"
)

// DefaultOutBase is where heatmap.geojson and heatmap.html are written.
const DefaultOutBase = "data/analytics/heatmap"

// Options configures a heatmap run.
type Options struct {
	Dir     string  // detail files dir; defaults to data/__Details
	CellDeg float64 // grid cell size in degrees; defaults to heatmap.DefaultCellDeg
	OutBase string  // output path without extension; defaults to DefaultOutBase
	Title   string  // map/page title
	Workers int     // file read concurrency; defaults to GOMAXPROCS
}

// Result describes what a run produced.
type Result struct {
	GeoJSONPath string
	HTMLPath    string
	Grid        heatmap.Grid
	Stats       analytics.Stats
}

// RunFileHeatmap reads product locations from detail JSON files, aggregates them
// into a grid, and writes <OutBase>.geojson plus a self-contained <OutBase>.html.
func RunFileHeatmap(opts Options) (Result, error) {
	if opts.OutBase == "" {
		opts.OutBase = DefaultOutBase
	}
	if opts.Title == "" {
		opts.Title = "Marketplace product heat map"
	}

	src := &source.FileSource{Dir: opts.Dir, Workers: opts.Workers}
	grid := heatmap.Aggregate(src, heatmap.Options{CellDeg: opts.CellDeg})

	geojson, err := grid.GeoJSON()
	if err != nil {
		return Result{}, fmt.Errorf("build geojson: %w", err)
	}
	html, err := grid.HTML(opts.Title)
	if err != nil {
		return Result{}, fmt.Errorf("build html: %w", err)
	}

	if dir := filepath.Dir(opts.OutBase); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Result{}, fmt.Errorf("create output dir: %w", err)
		}
	}

	geojsonPath := opts.OutBase + ".geojson"
	htmlPath := opts.OutBase + ".html"

	if err := os.WriteFile(geojsonPath, geojson, 0o644); err != nil {
		return Result{}, fmt.Errorf("write geojson: %w", err)
	}
	if err := os.WriteFile(htmlPath, html, 0o644); err != nil {
		return Result{}, fmt.Errorf("write html: %w", err)
	}

	return Result{
		GeoJSONPath: geojsonPath,
		HTMLPath:    htmlPath,
		Grid:        grid,
		Stats:       src.Stats(),
	}, nil
}
