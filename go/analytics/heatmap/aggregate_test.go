package heatmap

import (
	"testing"

	"github.com/romayengineer/fme/go/analytics"
)

type sliceSource struct{ points []analytics.GeoPoint }

func (s sliceSource) ForEach(fn func(analytics.GeoPoint) bool) int {
	n := 0
	for _, p := range s.points {
		if !fn(p) {
			break
		}
		n++
	}
	return n
}

func TestAggregateBinsPointsIntoGrid(t *testing.T) {
	src := sliceSource{points: []analytics.GeoPoint{
		{Lat: -34.605, Lon: -58.405}, // cell (x=-5841, y=-3461)
		{Lat: -34.603, Lon: -58.403}, // same cell as above
		{Lat: -34.70, Lon: -58.50},   // different cell
	}}

	grid := Aggregate(src, Options{CellDeg: 0.01})

	if grid.Total != 3 {
		t.Fatalf("Total = %d, want 3", grid.Total)
	}
	if len(grid.Cells) != 2 {
		t.Fatalf("len(Cells) = %d, want 2", len(grid.Cells))
	}
	if grid.Cells[0].Count != 2 {
		t.Errorf("top cell count = %d, want 2 (sorted desc)", grid.Cells[0].Count)
	}
	if grid.MaxCount != 2 {
		t.Errorf("MaxCount = %d, want 2", grid.MaxCount)
	}
}

func TestAggregateCellBounds(t *testing.T) {
	src := sliceSource{points: []analytics.GeoPoint{{Lat: 1.23, Lon: -4.56}}}
	grid := Aggregate(src, Options{CellDeg: 0.5})

	c := grid.Cells[0]
	if c.MinLon != -5.0 || c.MaxLon != -4.5 {
		t.Errorf("lon bounds = [%v,%v], want [-5,-4.5]", c.MinLon, c.MaxLon)
	}
	if c.MinLat != 1.0 || c.MaxLat != 1.5 {
		t.Errorf("lat bounds = [%v,%v], want [1,1.5]", c.MinLat, c.MaxLat)
	}
}

func TestAggregateDefaultCell(t *testing.T) {
	grid := Aggregate(sliceSource{}, Options{})
	if grid.Options.CellDeg != DefaultCellDeg {
		t.Errorf("CellDeg = %v, want default %v", grid.Options.CellDeg, DefaultCellDeg)
	}
}

func TestAggregateBounds(t *testing.T) {
	src := sliceSource{points: []analytics.GeoPoint{
		{Lat: -10, Lon: 20},
		{Lat: 5, Lon: -30},
		{Lat: 0.5, Lon: 0.5},
	}}
	grid := Aggregate(src, Options{CellDeg: 1})

	want := Bounds{MinLat: -10, MaxLat: 5, MinLon: -30, MaxLon: 20}
	if grid.Bounds != want {
		t.Errorf("Bounds = %+v, want %+v", grid.Bounds, want)
	}
}

func TestAggregateEmpty(t *testing.T) {
	grid := Aggregate(sliceSource{}, Options{CellDeg: 0.01})
	if grid.Total != 0 || len(grid.Cells) != 0 || grid.MaxCount != 0 {
		t.Errorf("empty source produced %+v", grid)
	}
}

func TestIntensity(t *testing.T) {
	grid := Grid{MaxCount: 4}
	if got := grid.Intensity(Cell{Count: 2}); got != 0.5 {
		t.Errorf("Intensity = %v, want 0.5", got)
	}
	if got := (Grid{}).Intensity(Cell{Count: 1}); got != 0 {
		t.Errorf("Intensity with no max = %v, want 0", got)
	}
}
