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

func WriteRandomJsonFile(body []byte) error {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(1000000)

	filename := filepath.Join("data", fmt.Sprintf("response_%d_%06d.json", timestamp, random))

	if err := os.WriteFile(filename, body, 0644); err != nil {
		return err
	}

	return nil
}

func WriteRandomJsonFileIndented(jsonData any, body []byte) error {
	indented, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return WriteRandomJsonFile(body)
	}
	return WriteRandomJsonFile(indented)
}

func WriteJsonResponse(body []byte) error {
	// make sure the first byte is { (open curly brakets)
	if body[0] != '{' {
		return nil
	}

	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		WriteRandomJsonFileIndented(jsonData, body)
	}

	var lineData interface{}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			WriteRandomJsonFileIndented(lineData, lineByte)
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
			}
		}()
	})

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()

}
