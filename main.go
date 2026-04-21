package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

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
					fmt.Printf("Response body: %s\n", string(body))
				}
			}
		}()
	})

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()

}
