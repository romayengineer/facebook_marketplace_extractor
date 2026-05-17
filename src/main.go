package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GetKey(data any, path string) any {
	keys := strings.Split(path, ".")

	current := data
	for _, key := range keys {
		dataMap, ok := current.(map[string]any)
		if !ok {
			// err := fmt.Errorf("cannot access key %q: not a map", key)
			// fmt.Println(err)
			return nil
		}

		value, ok := dataMap[key]
		if !ok {
			// err := fmt.Errorf("key %q not found", key)
			// fmt.Println(err)
			return nil
		}
		current = value
	}

	return current
}

func ExtractJsonFromBody(body []byte) ([]any, error) {
	jsonDatas := []any{}
	// make sure the first byte is { (open curly brakets)
	if body[0] != '{' {
		return jsonDatas, nil
	}

	var jsonData any
	if err := json.Unmarshal(body, &jsonData); err == nil {
		jsonDatas = append(jsonDatas, jsonData)
		return jsonDatas, nil
	}

	var lineData any
	for line := range strings.SplitSeq(string(body), "\n") {
		if line[0] != '{' {
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

func Begin() (ContextWrapperInterface, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error NewConfig: %v", err)
	}

	playwrightWrapper, err := NewPlaywrightWrapper()
	if err != nil {
		return nil, fmt.Errorf("error NewPlaywrightWrapper: %v", err)
	}

	browser, err := playwrightWrapper.NewBrowser(false)
	if err != nil {
		return nil, fmt.Errorf("error NewBrowser: %v", err)
	}

	facebookScrapper := NewFacebookScrapper(browser)

	ctx, err := facebookScrapper.Login(config.UserCredentials)
	if err != nil {
		return nil, fmt.Errorf("error Login: %v", err)
	}

	// productExtractors := NewProductExtractors()

	SetContextEventHandlers(ctx)

	return ctx, nil
}

func SearchProducts() {
	ctx, err := Begin()
	if err != nil {
		LogError0("SearchProducts", "Error in Begin", "error", err)
		os.Exit(1)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()
}

func GetDetails() {
	ctx, err := Begin()
	if err != nil {
		LogError0("GetDetails", "Error in Begin", "error", err)
		os.Exit(1)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)

	ForEachDetail(func(filePath string, jsonData any) bool {
		PriceAmount := GetKey(jsonData, "PriceAmount")
		if PriceAmount != nil {
			return true
		}

		productId := GetKey(jsonData, "ID")
		if productId == nil {
			return true
		}

		pages.GoToProduct(productId.(string))
		time.Sleep(3 * time.Second)

		return true
	}, false)

	WaitingForInput()
}

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

func GetTimestamp(filePath string) (int, error) {
	lastFilePathParts := strings.SplitN(filepath.Base(filePath), "_", 3)
	lastTimestampStr := lastFilePathParts[1]
	lastTimestamp, err := strconv.Atoi(lastTimestampStr)
	if err != nil {
		LogError0("GetTimestamp", "cound not get timestamp", "filePath", filePath)
		return 0, fmt.Errorf("cound not get timestamp %s\n", filePath)
	}
	return lastTimestamp, nil
}

func ProcessData(startAtTimestamp int) (int, error) {
	var lastFilePath string
	var filesProcessedCounter int
	var filesDeletedCounter int

	productExtractors := NewProductExtractors()
	ForEachResponse(func(filePath string, jsonData any) bool {
		lastFilePath = filePath

		fileTimestamp, err := GetTimestamp(filePath)
		if err != nil {
			return true
		}

		if fileTimestamp < startAtTimestamp {
			return true
		}

		filesProcessedCounter += 1

		for _, extractor := range productExtractors.extractors {
			product, _ := extractor.extractor(jsonData)
			if hasAny := SaveProductsIfAny(product); hasAny == true {
				return true
			}
		}

		LogDebug0("ProcessData", "no product found, deleting file", "path", filePath)
		if err := os.Remove(filePath); err != nil {
			LogError0("ProcessData", "error deleting file", "path", filePath, "error", err)
		}

		filesDeletedCounter += 1

		return true

	}, true)

	lastTimestamp, err := GetTimestamp(lastFilePath)
	if err != nil {
		return lastTimestamp, err
	}

	LogInfo0("ProcessData", "all files processed", "lastTimestamp", lastTimestamp, "filesProcessedCounter", filesProcessedCounter, "filesDeletedCounter", filesDeletedCounter)

	return lastTimestamp, nil
}

func main() {
	flags := NewFlags()

	switch flags.action {
	case "search":
		SearchProducts()
	case "process_data":
		ProcessData(0)
	case "get_details":
		GetDetails()
	default:
		LogError0("main", "Unknown action", "action", flags.action)
		os.Exit(1)
	}
}
