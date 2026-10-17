package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeDetail(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestRunFileHeatmapWritesArtifacts(t *testing.T) {
	dir := t.TempDir()
	writeDetail(t, dir, "detail_1.json", `{"LocationLatitud": -34.605, "LocationLongitude": -58.405}`)
	writeDetail(t, dir, "detail_2.json", `{"LocationLatitud": -34.603, "LocationLongitude": -58.403}`)
	writeDetail(t, dir, "detail_3.json", `{"LocationLatitud": null, "LocationLongitude": null}`)

	outBase := filepath.Join(t.TempDir(), "nested", "heatmap")
	res, err := RunFileHeatmap(Options{Dir: dir, CellDeg: 0.01, OutBase: outBase})
	if err != nil {
		t.Fatalf("RunFileHeatmap: %v", err)
	}

	if res.Stats.Valid != 2 || res.Stats.Skipped != 1 {
		t.Errorf("stats = %+v, want Valid:2 Skipped:1", res.Stats)
	}
	if len(res.Grid.Cells) != 1 || res.Grid.Cells[0].Count != 2 {
		t.Errorf("grid cells = %+v, want one cell with 2", res.Grid.Cells)
	}

	geo, err := os.ReadFile(res.GeoJSONPath)
	if err != nil {
		t.Fatalf("read geojson: %v", err)
	}
	var fc map[string]any
	if err := json.Unmarshal(geo, &fc); err != nil {
		t.Fatalf("geojson is not valid JSON: %v", err)
	}
	if fc["type"] != "FeatureCollection" {
		t.Errorf("geojson type = %v", fc["type"])
	}

	html, err := os.ReadFile(res.HTMLPath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	if len(html) == 0 {
		t.Error("html is empty")
	}
}
