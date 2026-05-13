package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func WriteRandomJsonFile(prefix string, body []byte) error {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(1000000)

	filename := filepath.Join("data", fmt.Sprintf("%s_%d_%06d.json", prefix, timestamp, random))

	if err := os.WriteFile(filename, body, 0644); err != nil {
		return err
	}

	return nil
}

func WriteRandomJsonFileIndented(prefix string, jsonData any) error {
	indented, _ := json.MarshalIndent(jsonData, "", "  ")
	return WriteRandomJsonFile(prefix, indented)
}
