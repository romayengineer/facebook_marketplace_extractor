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

func WriteResponse(body []byte) error {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(1000000)

	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	var extension string

	// Validate JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		extension = "txt"
	} else {
		extension = "json"
	}

	filename := filepath.Join(dataDir, fmt.Sprintf("response_%d_%d.%s", timestamp, random, extension))

	if err := os.WriteFile(filename, body, 0644); err != nil {
		return err
	}

	fmt.Printf("Wrote response to: %s\n", filename)
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
			if response.Ok() == false {
				return
			}
			request := response.Request()
			url := request.URL()
			if strings.Contains(url, "/api/graphql") {
				body, err := response.Body()
				if err == nil {
					WriteResponse(body)
				}
			}
		}()
	})

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()

}
