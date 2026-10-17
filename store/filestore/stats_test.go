package filestore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeDetail(t *testing.T, dir string, id string, description any) {
	t.Helper()
	record := map[string]any{"ID": id, "Description": description}
	body, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "detail_"+id+".json"), body, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func TestComputeDescriptionStats(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	detailDir := filepath.Join(DataDir, "__Details")
	if err := os.MkdirAll(detailDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	writeDetail(t, detailDir, "1", "a real description")
	writeDetail(t, detailDir, "2", "")
	writeDetail(t, detailDir, "3", nil)
	writeDetail(t, detailDir, "4", nil)

	stats := ComputeDescriptionStats()

	if stats.Total != 4 {
		t.Errorf("Total = %d, want 4", stats.Total)
	}
	if stats.Empty != 2 {
		t.Errorf("Empty = %d, want 2", stats.Empty)
	}
	if stats.WithDescription != 2 {
		t.Errorf("WithDescription = %d, want 2", stats.WithDescription)
	}
	if got := stats.EmptyPercent(); got != 50 {
		t.Errorf("EmptyPercent = %v, want 50", got)
	}
	if got := stats.WithDescriptionPercent(); got != 50 {
		t.Errorf("WithDescriptionPercent = %v, want 50", got)
	}
}

func TestComputeDescriptionStatsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(DataDir, "__Details"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	stats := ComputeDescriptionStats()

	if stats.Total != 0 {
		t.Errorf("Total = %d, want 0", stats.Total)
	}
	if got := stats.EmptyPercent(); got != 0 {
		t.Errorf("EmptyPercent = %v, want 0", got)
	}
	if got := stats.WithDescriptionPercent(); got != 0 {
		t.Errorf("WithDescriptionPercent = %v, want 0", got)
	}
}
