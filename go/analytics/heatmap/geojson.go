package heatmap

import "encoding/json"

type geoFeatureCollection struct {
	Type     string       `json:"type"`
	BBox     []float64    `json:"bbox,omitempty"`
	Features []geoFeature `json:"features"`
}

type geoFeature struct {
	Type       string         `json:"type"`
	Geometry   geoGeometry    `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type geoGeometry struct {
	Type        string         `json:"type"`
	Coordinates [][][2]float64 `json:"coordinates"`
}

// GeoJSON renders the grid as an RFC7946 FeatureCollection with one square
// Polygon per non-empty cell. Each feature carries the cell count and its
// normalized intensity, so any map library (or QGIS / geojson.io) can render it.
func (g Grid) GeoJSON() ([]byte, error) {
	fc := geoFeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]geoFeature, 0, len(g.Cells)),
	}

	if g.Total > 0 {
		fc.BBox = []float64{g.Bounds.MinLon, g.Bounds.MinLat, g.Bounds.MaxLon, g.Bounds.MaxLat}
	}

	for _, c := range g.Cells {
		ring := [][2]float64{
			{c.MinLon, c.MinLat},
			{c.MaxLon, c.MinLat},
			{c.MaxLon, c.MaxLat},
			{c.MinLon, c.MaxLat},
			{c.MinLon, c.MinLat},
		}
		fc.Features = append(fc.Features, geoFeature{
			Type:     "Feature",
			Geometry: geoGeometry{Type: "Polygon", Coordinates: [][][2]float64{ring}},
			Properties: map[string]any{
				"count":     c.Count,
				"intensity": g.Intensity(c),
				"cell_x":    c.Key.X,
				"cell_y":    c.Key.Y,
			},
		})
	}

	return json.MarshalIndent(fc, "", "  ")
}
