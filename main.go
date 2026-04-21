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

	ctx.Route("**/*", func(route playwright.Route) {
		request := route.Request()
		url := request.URL()
		if strings.Contains(url, "/api/graphql") {
			fmt.Printf("Intercepted: %s %s\n", request.Method(), url)
		}

		// Continue request as-is
		route.Continue()
	})

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()

}
