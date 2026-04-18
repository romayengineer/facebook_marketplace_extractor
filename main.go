package main

import (
	"log"
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

	facebookScrapper.Login(config.UserCredentials)
}
