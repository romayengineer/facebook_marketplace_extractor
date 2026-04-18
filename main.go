package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	playwrightWrapper, err := NewPlaywrightWrapper()
	if err != nil {
		log.Fatalf("error in main: %v", err)
	}

	defer playwrightWrapper.Playwright.Stop()

	browser, err := playwrightWrapper.NewBrowser(false)
	if err != nil {
		log.Fatalf("error in main: %v", err)
	}

	defer browser.Close()

	// Create a new page
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}
	defer page.Close()

	// Navigate to Facebook
	if _, err = page.Goto("https://www.facebook.com"); err != nil {
		log.Fatalf("Could not goto www.facebook.com: %v", err)
	}

	fmt.Println("Successfully opened www.facebook.com in Chromium")

	time.Sleep(10 * time.Minute)
}
