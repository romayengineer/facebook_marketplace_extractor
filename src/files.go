package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func SaveProductsIfAny(products []MarketplaceItemDetails) bool {
	if len(products) > 0 {
		for _, product := range products {
			store := NewProductFileStore(product.ID.(string))
			store.Save(product)
		}
		return true
	}
	return false
}

func ProcessData(StartTimeToProcess int64) (int64, error) {
	var lastTimestamp int64
	var lastFilePath string
	var filesProcessedCounter int
	var filesDeletedCounter int
	var gotRateLimit bool

	productExtractors := NewProductExtractors()
	filesReadCounter := ForEachResponse(func(filePath string, jsonData map[string]any) bool {
		lastFilePath = filePath

		fileTimestamp, err := GetTimestamp(filePath)
		if err != nil {
			return true
		}

		if fileTimestamp < StartTimeToProcess {
			return true
		}

		filesProcessedCounter += 1

		for _, extractor := range productExtractors.extractors {
			product, _ := extractor.extractor(jsonData)
			if hasAny := SaveProductsIfAny(product); hasAny == true {
				return true
			}
		}

		if IsErrorRateLimit(jsonData) {
			gotRateLimit = true
		}

		LogDebug0("ProcessData", "no product found, deleting file", "path", filePath)
		if err := os.Remove(filePath); err != nil {
			LogError0("ProcessData", "error deleting file", "path", filePath, "error", err)
		}

		filesDeletedCounter += 1

		return true

	}, true)

	lastTimestamp, _ = GetTimestamp(lastFilePath)

	LogInfo0("ProcessData", "all files processed", "lastTimestamp", lastTimestamp, "filesProcessedCounter", filesProcessedCounter, "filesReadCounter", filesReadCounter, "filesDeletedCounter", filesDeletedCounter)

	if gotRateLimit {
		LogFatal("got rate limit")
	}

	return lastTimestamp, nil
}

func GetTimestamp(filePath string) (int64, error) {
	lastFilePathParts := strings.SplitN(filepath.Base(filePath), "_", 3)
	if len(lastFilePathParts) < 2 {
		LogWarn0("GetTimestamp", "filePath does not have time", "filePath", filePath)
		return 0, fmt.Errorf("filePath does not have time")
	}
	lastTimestampStr := lastFilePathParts[1]
	lastTimestamp, err := strconv.ParseInt(lastTimestampStr, 10, 64)
	if err != nil {
		LogError0("GetTimestamp", "cound not get timestamp", "filePath", filePath)
		return 0, fmt.Errorf("cound not get timestamp %s\n", filePath)
	}
	return lastTimestamp, nil
}

func GetKey(data any, path string) (any, bool) {
	keys := strings.Split(path, ".")

	current := data
	for _, key := range keys {
		dataMap, ok := current.(map[string]any)
		if !ok {
			// err := fmt.Errorf("cannot access key %q: not a map", key)
			// fmt.Println(err)
			return current, false
		}

		value, ok := dataMap[key]
		if !ok {
			// err := fmt.Errorf("key %q not found", key)
			// fmt.Println(err)
			return dataMap, false
		}
		current = value
	}

	return current, true
}

func ExtractJsonFromBody(body []byte) ([]any, error) {
	jsonDatas := []any{}
	// make sure the first byte is { (open curly brakets)
	if len(body) == 0 || body[0] != '{' {
		return jsonDatas, nil
	}

	var jsonData any
	if err := json.Unmarshal(body, &jsonData); err == nil {
		jsonDatas = append(jsonDatas, jsonData)
		return jsonDatas, nil
	}

	var lineData any
	for line := range strings.SplitSeq(string(body), "\n") {
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			jsonDatas = append(jsonDatas, lineData)
		}
	}

	return jsonDatas, nil
}

func WriteJsonResponse(jsonDatas []any, friendly_name string) (int, error) {
	jsonCounter := 0
	var err error
	for _, jsonData := range jsonDatas {
		if err = WriteRandomJsonFileIndented("response", friendly_name, jsonData); err != nil {
			return jsonCounter, err
		}
		jsonCounter += 1
	}

	return jsonCounter, nil
}

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

func TimeNow() int64 {
	return time.Now().UnixNano()
}

func TimeYesterday() int64 {
	return time.Now().AddDate(0, 0, -1).UnixNano()
}

func WriteRandomJsonFile(prefix string, friendly_name string, body []byte) error {
	timestamp := TimeNow()
	random := rand.Intn(1000000)

	fileDir := filepath.Join(DataDir, friendly_name)

	filename := filepath.Join(fileDir, fmt.Sprintf("%s_%d_%06d.json", prefix, timestamp, random))

	WriteFileAndDirs(filename, body, 0644)

	return nil
}

func WriteRandomJsonFileIndented(prefix string, friendly_name string, jsonData any) error {
	indented, _ := json.MarshalIndent(jsonData, "", "  ")
	return WriteRandomJsonFile(prefix, friendly_name, indented)
}

func GetFilePaths(prefix string, sortit bool) []string {

	filePaths := []string{}

	err := filepath.WalkDir(DataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// skip archived folder
		if strings.Contains(path, "archived") {
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
		LogError0("GetFilePaths", "Error reading data directory", "error", err)
		os.Exit(1)
	}

	if sortit {
		// if sort is true sort filePaths in ascendant order
		sort.Slice(filePaths, func(i, j int) bool {
			return filepath.Base(filePaths[i]) < filepath.Base(filePaths[j])
		})
	} else {
		// else sort filePaths in random order
		rand.Shuffle(len(filePaths), func(i, j int) { filePaths[i], filePaths[j] = filePaths[j], filePaths[i] })
	}

	return filePaths
}

func ForEachJsonInData(prefix string, process func(filePath string, jsonData map[string]any) bool, sortit bool) int {
	// open and read all files in data folder that start with response and end in .json

	filesReadCounter := 0

	filePaths := GetFilePaths(prefix, sortit)

	for _, filePath := range filePaths {

		// fmt.Printf("%s\n", filePath)

		body, err := os.ReadFile(filePath)
		if err != nil {
			LogError0("ForEachJsonInData", "Error reading file", "path", filePath, "error", err)
			continue
		}

		var jsonData map[string]any
		if err := json.Unmarshal(body, &jsonData); err != nil {
			LogError0("ForEachJsonInData", "Error parsing JSON", "path", filePath, "error", err)
			continue
		}

		shouldContinue := process(filePath, jsonData)

		filesReadCounter++

		if !shouldContinue {
			return filesReadCounter
		}
	}

	return filesReadCounter

}

func ForEachResponse(process func(filePath string, jsonData map[string]any) bool, sortit bool) int {
	return ForEachJsonInData("response_", process, sortit)
}

func ForEachDetail(process func(filePath string, jsonData map[string]any) bool, sortit bool) int {
	return ForEachJsonInData("detail_", process, sortit)
}

func FillEmpty() {

	ForEachDetail(func(filePath string, jsonData map[string]any) bool {
		productUrl, _ := GetKey(jsonData, "URL")
		if productUrl != nil {
			return true
		}

		productId, _ := GetKey(jsonData, "ID")
		if productId == nil {
			return true
		}

		productIdStr, ok := productId.(string)

		if !ok {
			return true
		}

		productUrl = ProductIDTolink(productIdStr)

		jsonData["URL"] = productUrl

		store := NewProductFileStore(productIdStr)
		store.Save(ToMarketplaceItemDetails(jsonData))

		return true
	}, false)
}
