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

func WriteJsonResponse(body []byte) error {
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		return WriteRandomJsonFile(body)
	}

	var lineData interface{}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			WriteRandomJsonFile(lineByte)
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
			if response.Ok() == false {
				return
			}
			request := response.Request()
			url := request.URL()
			if strings.Contains(url, "/api/graphql") {
				body, err := response.Body()
				if err == nil {
					WriteJsonResponse(body)
				}
			}
		}()
	})

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()

}
