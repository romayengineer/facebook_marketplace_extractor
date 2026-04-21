package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
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

func WriteRandomJsonFileIndented(prefix string, body []byte, jsonData any) error {
	indented, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return WriteRandomJsonFile(prefix, body)
	}
	return WriteRandomJsonFile(prefix, indented)
}

func GetKey(data any, path string) (any, error) {
	keys := strings.Split(path, ".")

	current := data
	for _, key := range keys {
		dataMap, ok := current.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("cannot access key %q: not a map", key)
			fmt.Println(err)
			return nil, err
		}

		value, ok := dataMap[key]
		if !ok {
			err := fmt.Errorf("key %q not found", key)
			fmt.Println(err)
			return nil, err
		}
		current = value
	}

	return current, nil
}

func GetProductDetails(data any) (any, error) {
	path := "data.viewer.marketplace_product_details_page.target"
	return GetKey(data, path)
}

func GetProductsFromSearch(data any) (any, error) {
	path := "data.marketplace_search.feed_units.edges"
	return GetKey(data, path)
}

func WriteJsonResponse(body []byte) error {
	// make sure the first byte is { (open curly brakets)
	if body[0] != '{' {
		return nil
	}

	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		return WriteRandomJsonFileIndented("response", body, jsonData)
	}

	var lineData interface{}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] != '{' {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			WriteRandomJsonFileIndented("response", lineByte, lineData)
		}
	}

	return nil

}

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatalf("error NewConfig: %v", err)
	}

	playwrightWrapper, err := NewPlaywrightWrapper()
	if err != nil {
		log.Fatalf("error NewPlaywrightWrapper: %v", err)
	}

	defer playwrightWrapper.Stop()

	browser, err := playwrightWrapper.NewBrowser(false)
	if err != nil {
		log.Fatalf("error NewBrowser: %v", err)
	}
	defer browser.Close()

	facebookScrapper := NewFacebookScrapper(browser)

	ctx, err := facebookScrapper.Login(config.UserCredentials)
	if err != nil {
		log.Fatalf("error Login: %v", err)
	}
	defer ctx.Close()

	ctx.OnResponse(func(response playwright.Response) {
		go func() {
			response.Finished()
			body, err := response.Body()
			if err == nil {
				WriteJsonResponse(body)
			} else {
				fmt.Printf("Error OnResponse: %v", err)
			}
		}()
	})

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()

}
