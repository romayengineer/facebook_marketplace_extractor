// Package heatmap aggregates geographic points into a fixed lat/lon grid and
// renders that grid as a heat map (GeoJSON or a self-contained HTML page).
//
// The grid approach counts products per square cell instead of plotting every
// point: a single O(n) pass over the points and O(cells) memory, which scales to
// millions of listings without a spatial index.
package heatmap

// DefaultCellDeg is the default cell size in degrees (~1.1 km at the equator).
const DefaultCellDeg = 0.01

// Options controls how points are binned.
type Options struct {
	// CellDeg is the cell edge length in degrees. Values <= 0 fall back to
	// DefaultCellDeg.
	CellDeg float64
}

// withDefaults returns a copy with a usable CellDeg.
func (o Options) withDefaults() Options {
	if o.CellDeg <= 0 {
		o.CellDeg = DefaultCellDeg
	}
	return o
}

// CellKey identifies a grid cell by integer indices, which keeps map lookups
// allocation-free (no float keys, no geohash strings).
type CellKey struct {
	X int32 // longitude band
	Y int32 // latitude band
}

// Cell is one non-empty grid square.
type Cell struct {
	Key            CellKey
	Count          int
	MinLat, MinLon float64
	MaxLat, MaxLon float64
}

// Bounds is the bounding box of every aggregated point.
type Bounds struct {
	MinLat, MinLon float64
	MaxLat, MaxLon float64
}

// Grid is the aggregation result for a set of points.
type Grid struct {
	Options  Options
	Total    int // points aggregated
	MaxCount int // highest per-cell count (intensity denominator)
	Cells    []Cell
	Bounds   Bounds
}
