package filestore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveProductsFromMergesDescription(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	detailDir := filepath.Join(DataDir, "__Details")
	if err := os.MkdirAll(detailDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	existing := map[string]any{"ID": "123", "Title": "a listing", "Description": nil}
	body, err := json.Marshal(existing)
	if err != nil {
		t.Fatalf("marshal existing: %v", err)
	}
	if err := os.WriteFile(filepath.Join(detailDir, "detail_123.json"), body, 0644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	pdp := map[string]any{
		"data": map[string]any{
			"viewer": map[string]any{
				"marketplace_product_details_page": map[string]any{
					"target": map[string]any{
						"id":                   "123",
						"redacted_description": map[string]any{"text": "hello"},
					},
				},
			},
		},
	}

	if saved := SaveProductsFrom([]any{pdp}); saved != 1 {
		t.Fatalf("SaveProductsFrom = %d, want 1", saved)
	}

	got, err := os.ReadFile(filepath.Join(detailDir, "detail_123.json"))
	if err != nil {
		t.Fatalf("read detail: %v", err)
	}

	var record map[string]any
	if err := json.Unmarshal(got, &record); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if record["Description"] != "hello" {
		t.Errorf("Description = %v, want hello", record["Description"])
	}
	if record["Title"] != "a listing" {
		t.Errorf("Title = %v, want a listing (existing field preserved)", record["Title"])
	}
}

func TestSaveProductsFromNoProduct(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if saved := SaveProductsFrom([]any{map[string]any{"unrelated": true}}); saved != 0 {
		t.Errorf("SaveProductsFrom = %d, want 0", saved)
	}
}
