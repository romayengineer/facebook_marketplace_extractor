package heatmap

import "html/template"

// templateData is the shared contract every map template renders. Templates pick
// the fields they need (Token is only used by providers that require a key).
type templateData struct {
	Title    string
	Total    int
	Cells    int
	MaxCount int
	CellDeg  float64
	GeoJSON  template.JS
	Token    string
}

// HTML renders the grid with the default map. Use RenderHTML to pick a map.
func (g Grid) HTML(title string) ([]byte, error) {
	return g.RenderHTML(DefaultMapID, MapOptions{Title: title})
}
