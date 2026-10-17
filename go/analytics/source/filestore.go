// Package source provides PointSource adapters over the different places FME
// stores scraped listings.
package source

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/romayengineer/fme/go/analytics"
)

// DefaultDir is where product detail files live.
const DefaultDir = "data/__Details"

// FileSource reads product locations from data/__Details/detail_*.json, the same
// files the rest of FME materializes, so heat maps work without the PocketBase
// database. File reads are parallelized across workers; results are replayed in
// filename order so aggregation stays deterministic.
type FileSource struct {
	Dir     string // defaults to DefaultDir
	Workers int    // defaults to GOMAXPROCS

	mu    sync.Mutex
	stats analytics.Stats
}

// detailRecord decodes only the fields we need. The coordinate fields are
// pointers so JSON null and malformed values are simply skipped.
type detailRecord struct {
	LocationLatitud   *float64 `json:"LocationLatitud"`
	LocationLongitude *float64 `json:"LocationLongitude"`
}

// Stats returns coverage counters for the last ForEach call.
func (s *FileSource) Stats() analytics.Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

// ForEach implements analytics.PointSource.
func (s *FileSource) ForEach(fn func(analytics.GeoPoint) bool) int {
	dir := s.Dir
	if dir == "" {
		dir = DefaultDir
	}
	workers := s.Workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	paths := detailFiles(dir)
	points := s.readAll(paths, workers)

	emitted := 0
	for _, p := range points {
		if !fn(p) {
			break
		}
		emitted++
	}
	return emitted
}

// readAll reads the given files concurrently and returns the valid points in the
// same order as paths.
func (s *FileSource) readAll(paths []string, workers int) []analytics.GeoPoint {
	results := make([][]analytics.GeoPoint, len(paths))
	scanned := make([]bool, len(paths))

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = readPoints(paths[i])
				scanned[i] = true
			}
		}()
	}
	for i := range paths {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	s.mu.Lock()
	s.stats = analytics.Stats{}
	points := make([]analytics.GeoPoint, 0, len(paths))
	for i := range paths {
		if !scanned[i] {
			continue
		}
		s.stats.Scanned++
		points = append(points, results[i]...)
	}
	s.stats.Valid = len(points)
	s.stats.Skipped = s.stats.Scanned - s.stats.Valid
	s.mu.Unlock()

	return points
}

// detailFiles lists data/__Details entries in deterministic (sorted) order.
func detailFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "detail_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, name))
	}
	sort.Strings(paths)
	return paths
}

// readPoints parses one detail file and returns its point, if the coordinates
// are usable. A missing/invalid file yields no points.
func readPoints(path string) []analytics.GeoPoint {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rec detailRecord
	if err := json.Unmarshal(body, &rec); err != nil {
		return nil
	}
	if rec.LocationLatitud == nil || rec.LocationLongitude == nil {
		return nil
	}
	lat, lon := *rec.LocationLatitud, *rec.LocationLongitude
	if !validCoord(lat, lon) {
		return nil
	}
	return []analytics.GeoPoint{{Lat: lat, Lon: lon}}
}

// validCoord rejects out-of-range values and Facebook's (0,0) "unknown"
// placeholder, which would otherwise pile every missing location onto the map at
// the Gulf of Guinea.
func validCoord(lat, lon float64) bool {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return false
	}
	return lat != 0 || lon != 0
}
