package main

import (
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"
)

func main() {
	// Install browsers if needed
	err := playwright.Install()
	if err != nil {
		log.Fatalf("Could not install browser drivers: %v", err)
	}

	// Start playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start playwright: %v", err)
	}
	defer pw.Stop()

	// Launch chromium browser
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("Could not launch chromium: %v", err)
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
}
