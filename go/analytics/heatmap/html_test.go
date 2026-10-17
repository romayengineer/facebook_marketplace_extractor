package heatmap

import (
	"bytes"
	"testing"

	"github.com/romayengineer/fme/go/analytics"
)

func TestHTMLUsesNonBlockingBasemap(t *testing.T) {
	src := sliceSource{points: []analytics.GeoPoint{{Lat: -34.6, Lon: -58.4}}}
	grid := Aggregate(src, Options{CellDeg: 0.01})

	html, err := grid.HTML("test")
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}

	if !bytes.Contains(html, []byte("basemaps.cartocdn.com")) {
		t.Error("expected CARTO basemap tiles")
	}
	if bytes.Contains(html, []byte("tile.openstreetmap.org")) {
		t.Error("must not use OSM's volunteer tile server (blocks browser apps)")
	}
	if !bytes.Contains(html, []byte("FeatureCollection")) {
		t.Error("expected embedded GeoJSON")
	}
}
