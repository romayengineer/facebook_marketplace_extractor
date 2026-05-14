package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
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

func ForEachJsonInData(prefix string, process func(filePath string, jsonData any), sortit bool) {
	// open and read all files in data folder that start with response and end in .json

	filePaths := []string{}

	err := filepath.WalkDir("data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileName := d.Name()

		if !strings.HasPrefix(fileName, prefix) || !strings.HasSuffix(fileName, ".json") {
			return nil
		}

		filePaths = append(filePaths, path)

		return nil
	})

	if err != nil {
		log.Fatalf("error reading data directory: %v", err)
	}

	if sortit {
		// if sort is true sort filePaths in ascendant order
		sort.Strings(filePaths)
	} else {
		// else sort filePaths in random order
		rand.Shuffle(len(filePaths), func(i, j int) { filePaths[i], filePaths[j] = filePaths[j], filePaths[i] })
	}

	for _, filePath := range filePaths {

		// fmt.Printf("%s\n", filePath)

		body, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("error reading file %s: %v", filePath, err)
			continue
		}

		var jsonData any
		if err := json.Unmarshal(body, &jsonData); err != nil {
			log.Printf("error parsing JSON from %s: %v", filePath, err)
			continue
		}

		process(filePath, jsonData)
	}

}

func ForEachResponse(process func(filePath string, jsonData any), sortit bool) {
	ForEachJsonInData("response_", process, sortit)
}

func ForEachDetail(process func(filePath string, jsonData any), sortit bool) {
	ForEachJsonInData("detail_", process, sortit)
}
