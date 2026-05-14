package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteFileAndDirs(name string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(name, data, perm); err != nil {
		if strings.Contains(err.Error(), "The system cannot find the path specified.") {
			nameDir := filepath.Dir(name)
			if err := os.MkdirAll(nameDir, 0755); err != nil {
				return err
			}
			if err := os.WriteFile(name, data, perm); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func WriteRandomJsonFile(prefix string, friendly_name string, body []byte) error {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(1000000)

	fileDir := filepath.Join("data", friendly_name)

	filename := filepath.Join(fileDir, fmt.Sprintf("%s_%d_%06d.json", prefix, timestamp, random))

	WriteFileAndDirs(filename, body, 0644)

	return nil
}

func WriteRandomJsonFileIndented(prefix string, friendly_name string, jsonData any) error {
	indented, _ := json.MarshalIndent(jsonData, "", "  ")
	return WriteRandomJsonFile(prefix, friendly_name, indented)
}
