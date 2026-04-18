package main

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type PlaywrightWrapperInterface interface {
	Run() error
	NewBrowser(headless bool) playwright.Browser
}

type PlaywrightWrapper struct {
	Playwright *playwright.Playwright
}

func NewPlaywrightWrapper() (*PlaywrightWrapper, error) {
	playwrightWrapper := PlaywrightWrapper{}
	err := playwrightWrapper.Run()
	if err != nil {
		return &playwrightWrapper, fmt.Errorf("Could not create PlaywrightWrapper: %v", err)
	}
	return &playwrightWrapper, nil
}

func (pw *PlaywrightWrapper) Run() error {
	// Install only Chromium browser
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		return fmt.Errorf("Could not install browser drivers: %v", err)
	}

	playwright, err := playwright.Run()
	pw.Playwright = playwright

	if err != nil {
		return fmt.Errorf("Could not start playwright: %v", err)
	}

	return err
}

func (pw *PlaywrightWrapper) NewBrowser(headless bool) (playwright.Browser, error) {
	// Launch chromium browser
	browser, err := pw.Playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
	})
	if err != nil {
		return browser, fmt.Errorf("Could not launch chromium: %v", err)
	}

	return browser, nil
}
