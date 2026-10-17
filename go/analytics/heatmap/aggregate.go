package heatmap

import (
	"math"
	"sort"

	"github.com/romayengineer/fme/go/analytics"
)

// Aggregate bins every point from src into a fixed lat/lon grid.
//
// It performs a single pass over the points, accumulating counts into a map
// keyed by integer cell indices, then flattens and sorts the cells by count
// (descending, ties broken by cell position) so the output is deterministic.
func Aggregate(src analytics.PointSource, opts Options) Grid {
	opts = opts.withDefaults()
	cell := opts.CellDeg

	counts := make(map[CellKey]int)
	grid := Grid{Options: opts}

	first := true
	src.ForEach(func(p analytics.GeoPoint) bool {
		key := CellKey{
			X: int32(math.Floor(p.Lon / cell)),
			Y: int32(math.Floor(p.Lat / cell)),
		}
		counts[key]++
		grid.Total++

		if first {
			grid.Bounds = Bounds{MinLat: p.Lat, MaxLat: p.Lat, MinLon: p.Lon, MaxLon: p.Lon}
			first = false
		} else {
			if p.Lat < grid.Bounds.MinLat {
				grid.Bounds.MinLat = p.Lat
			}
			if p.Lat > grid.Bounds.MaxLat {
				grid.Bounds.MaxLat = p.Lat
			}
			if p.Lon < grid.Bounds.MinLon {
				grid.Bounds.MinLon = p.Lon
			}
			if p.Lon > grid.Bounds.MaxLon {
				grid.Bounds.MaxLon = p.Lon
			}
		}
		return true
	})

	grid.Cells = make([]Cell, 0, len(counts))
	for key, count := range counts {
		minLon := float64(key.X) * cell
		minLat := float64(key.Y) * cell
		grid.Cells = append(grid.Cells, Cell{
			Key:    key,
			Count:  count,
			MinLat: minLat,
			MaxLat: minLat + cell,
			MinLon: minLon,
			MaxLon: minLon + cell,
		})
		if count > grid.MaxCount {
			grid.MaxCount = count
		}
	}

	sort.Slice(grid.Cells, func(i, j int) bool {
		if grid.Cells[i].Count != grid.Cells[j].Count {
			return grid.Cells[i].Count > grid.Cells[j].Count
		}
		if grid.Cells[i].Key.X != grid.Cells[j].Key.X {
			return grid.Cells[i].Key.X < grid.Cells[j].Key.X
		}
		return grid.Cells[i].Key.Y < grid.Cells[j].Key.Y
	})

	return grid
}

// Intensity returns the count normalized to [0,1] against the busiest cell.
func (g Grid) Intensity(c Cell) float64 {
	if g.MaxCount == 0 {
		return 0
	}
	return float64(c.Count) / float64(g.MaxCount)
}
