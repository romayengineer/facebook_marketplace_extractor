package main

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type BrowserWrapperInterface interface {
	NewPage() (playwright.Page, error)
	Close() error
}

type PlaywrightWrapperInterface interface {
	Run() error
	NewBrowser(headless bool) (BrowserWrapperInterface, error)
	Stop() error
}

type BrowserWrapper struct {
	Browser playwright.Browser
}

type PlaywrightWrapper struct {
	Playwright *playwright.Playwright
}

func NewBrowserWrapper(browser playwright.Browser) (BrowserWrapperInterface, error) {
	browserWrapper := BrowserWrapper{
		Browser: browser,
	}
	return &browserWrapper, nil
}

func (ww *BrowserWrapper) NewPage() (playwright.Page, error) {
	page, err := ww.Browser.NewPage()
	if err != nil {
		return page, fmt.Errorf("Could not create page: %v", err)
	}
	return page, nil
}

func (ww *BrowserWrapper) Close() error {
	return ww.Browser.Close()
}

func NewPlaywrightWrapper() (PlaywrightWrapperInterface, error) {
	playwrightWrapper := PlaywrightWrapper{}
	err := playwrightWrapper.Run()
	if err != nil {
		return &playwrightWrapper, fmt.Errorf("Could not create PlaywrightWrapper: %v", err)
	}
	return &playwrightWrapper, nil
}

func (pw *PlaywrightWrapper) Stop() error {
	return pw.Playwright.Stop()
}

func (pw *PlaywrightWrapper) Run() error {
	if pw.Playwright != nil {
		return fmt.Errorf("playwright already running")
	}

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

func (pw *PlaywrightWrapper) NewBrowser(headless bool) (BrowserWrapperInterface, error) {
	browser, err := pw.Playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
	})
	if err != nil {
		return nil, fmt.Errorf("Could not launch chromium: %v", err)
	}
	browserWrapper, nil := NewBrowserWrapper(browser)
	if err != nil {
		return browserWrapper, fmt.Errorf("Could not create BrowserWrapper: %v", err)
	}
	return browserWrapper, nil
}
