package heatmap

import (
	"bytes"
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

var htmlTemplate = template.Must(template.New("heatmap").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>{{.Title}}</title>
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
<style>
  html, body { margin: 0; height: 100%; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
  #map { position: absolute; inset: 0; }
  .panel { background: rgba(255,255,255,.92); padding: 10px 12px; border-radius: 8px; box-shadow: 0 1px 5px rgba(0,0,0,.3); font-size: 13px; line-height: 1.5; }
  .legend { position: absolute; bottom: 24px; right: 12px; z-index: 500; }
  .legend .bar { height: 12px; width: 180px; border-radius: 3px; background: linear-gradient(to right, rgb(255,255,204), rgb(128,0,38)); }
  .legend .labels { display: flex; justify-content: space-between; margin-top: 2px; color: #444; }
  .stats { position: absolute; top: 12px; right: 12px; z-index: 500; }
  .stats b { font-size: 15px; }
</style>
</head>
<body>
<div id="map"></div>
<div class="panel stats">
  <b>{{.Title}}</b><br />
  products: {{.Total}}<br />
  cells: {{.Cells}}<br />
  peak per cell: {{.MaxCount}}<br />
  cell: {{.CellDeg}}&deg;
</div>
<div class="panel legend">
  <div>products per cell</div>
  <div class="bar"></div>
  <div class="labels"><span>0</span><span>{{.MaxCount}}</span></div>
</div>
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
<script>
  var grid = {{.GeoJSON}};

  function lerp(a, b, t) { return a + (b - a) * t; }
  function colorFor(intensity) {
    var t = Math.max(0, Math.min(1, intensity));
    var r = Math.round(lerp(255, 128, t));
    var g = Math.round(lerp(255, 0, t));
    var b = Math.round(lerp(204, 38, t));
    return "rgb(" + r + "," + g + "," + b + ")";
  }

  var map = L.map("map");
  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map);

  var layer = L.geoJSON(grid, {
    style: function (feature) {
      var i = feature.properties.intensity;
      return {
        fillColor: colorFor(i),
        fillOpacity: 0.35 + 0.6 * Math.max(0, Math.min(1, i)),
        color: "#800026",
        weight: 0.5
      };
    },
    onEachFeature: function (feature, lyr) {
      var p = feature.properties;
      lyr.bindPopup("<b>" + p.count + "</b> products in this cell");
    }
  }).addTo(map);

  if (grid.features.length > 0) {
    map.fitBounds(layer.getBounds(), { padding: [30, 30] });
  } else {
    map.setView([0, 0], 2);
  }
</script>
</body>
</html>
`))

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
