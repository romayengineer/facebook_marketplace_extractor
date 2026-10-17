package heatmap

import (
	"encoding/json"
	"testing"

	"github.com/romayengineer/fme/go/analytics"
)

func TestGeoJSONShape(t *testing.T) {
	src := sliceSource{points: []analytics.GeoPoint{
		{Lat: -34.605, Lon: -58.405},
		{Lat: -34.603, Lon: -58.403},
	}}
	grid := Aggregate(src, Options{CellDeg: 0.01})

	raw, err := grid.GeoJSON()
	if err != nil {
		t.Fatalf("GeoJSON: %v", err)
	}

	var fc struct {
		Type     string `json:"type"`
		BBox     []float64
		Features []struct {
			Type     string `json:"type"`
			Geometry struct {
				Type        string         `json:"type"`
				Coordinates [][][2]float64 `json:"coordinates"`
			} `json:"geometry"`
			Properties map[string]any `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(raw, &fc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if fc.Type != "FeatureCollection" {
		t.Errorf("type = %q, want FeatureCollection", fc.Type)
	}
	if len(fc.Features) != 1 {
		t.Fatalf("features = %d, want 1", len(fc.Features))
	}

	f := fc.Features[0]
	if f.Geometry.Type != "Polygon" {
		t.Errorf("geometry type = %q, want Polygon", f.Geometry.Type)
	}
	ring := f.Geometry.Coordinates[0]
	if len(ring) != 5 {
		t.Fatalf("ring points = %d, want 5 (closed)", len(ring))
	}
	if ring[0] != ring[4] {
		t.Errorf("ring is not closed: %v vs %v", ring[0], ring[4])
	}
	if count := f.Properties["count"].(float64); count != 2 {
		t.Errorf("count = %v, want 2", count)
	}
	if intensity := f.Properties["intensity"].(float64); intensity != 1 {
		t.Errorf("intensity = %v, want 1", intensity)
	}
}
