package heatmap

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

type htmlData struct {
	Title    string
	Total    int
	Cells    int
	MaxCount int
	CellDeg  float64
	GeoJSON  template.JS
}

//go:embed template.html
var htmlTemplateSource string

var htmlTemplate = template.Must(template.New("heatmap").Parse(htmlTemplateSource))

// HTML renders a single self-contained page (Leaflet + the grid embedded as a
// JS literal) that shows the counts as a choropleth heat map. It only needs
// network access for the basemap tiles and the Leaflet CDN.
func (g Grid) HTML(title string) ([]byte, error) {
	geo, err := g.GeoJSON()
	if err != nil {
		return nil, fmt.Errorf("render geojson: %w", err)
	}

	var buf bytes.Buffer
	data := htmlData{
		Title:    title,
		Total:    g.Total,
		Cells:    len(g.Cells),
		MaxCount: g.MaxCount,
		CellDeg:  g.Options.CellDeg,
		GeoJSON:  template.JS(geo),
	}
	if err := htmlTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render html: %w", err)
	}
	return buf.Bytes(), nil
}
