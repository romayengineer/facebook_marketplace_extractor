package main

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type LocatorWrapperInterface interface {
	Fill(value string) error
	Click() error
	Locator(selector string) LocatorWrapperInterface
	GetByRole(role playwright.AriaRole) LocatorWrapperInterface
}

type PageWrapperInterface interface {
	Goto(url string) error
	Locator(selector string) LocatorWrapperInterface
	Close() error
	GetByRole(role playwright.AriaRole) LocatorWrapperInterface
}

type BrowserWrapperInterface interface {
	NewPage() (PageWrapperInterface, error)
	Close() error
}

type PlaywrightWrapperInterface interface {
	Run() error
	NewBrowser(headless bool) (BrowserWrapperInterface, error)
	Stop() error
}

type LocatorWrapper struct {
	locator playwright.Locator
}

type PageWrapper struct {
	Page playwright.Page
}

type BrowserWrapper struct {
	Browser playwright.Browser
}

type PlaywrightWrapper struct {
	Playwright *playwright.Playwright
}

func NewLocatorWrapper(locator playwright.Locator) LocatorWrapperInterface {
	locatorWrapper := LocatorWrapper{
		locator: locator,
	}
	return &locatorWrapper
}

func (lw *LocatorWrapper) GetByRole(role playwright.AriaRole) LocatorWrapperInterface {
	locator := lw.locator.GetByRole(role)
	locatorWrapper := NewLocatorWrapper(locator)
	return locatorWrapper
}

func (lw *LocatorWrapper) Locator(selector string) LocatorWrapperInterface {
	locator := lw.locator.Locator(selector)
	locatorWrapper := NewLocatorWrapper(locator)
	return locatorWrapper
}

func (lw *LocatorWrapper) Fill(value string) error {
	err := lw.locator.Fill(value)
	if err != nil {
		return fmt.Errorf("error on fill: %v", err)
	}
	return nil
}

func (lw *LocatorWrapper) Click() error {
	err := lw.locator.Click()
	if err != nil {
		return fmt.Errorf("error on click: %v", err)
	}
	return nil
}

func NewPageWrapper(page playwright.Page) PageWrapperInterface {
	pageWrapper := PageWrapper{
		Page: page,
	}
	return &pageWrapper
}

func (pw *PageWrapper) GetByRole(role playwright.AriaRole) LocatorWrapperInterface {
	locator := pw.Page.GetByRole(role)
	return NewLocatorWrapper(locator)
}

func (pw *PageWrapper) Goto(url string) error {
	_, err := pw.Page.Goto(url)
	if err != nil {
		return fmt.Errorf("Could not goto %s: %v", url, err)
	}
	return nil
}

func (pw *PageWrapper) Locator(selector string) LocatorWrapperInterface {
	locator := pw.Page.Locator(selector)
	locatorWrapper := NewLocatorWrapper(locator)
	return locatorWrapper
}

func (pw *PageWrapper) Close() error {
	return pw.Page.Close()
}

func NewBrowserWrapper(browser playwright.Browser) (BrowserWrapperInterface, error) {
	browserWrapper := BrowserWrapper{
		Browser: browser,
	}
	return &browserWrapper, nil
}

func (ww *BrowserWrapper) NewPage() (PageWrapperInterface, error) {
	page, err := ww.Browser.NewPage()
	pageWrapper := NewPageWrapper(page)
	if err != nil {
		return pageWrapper, fmt.Errorf("Could not create page: %v", err)
	}
	return pageWrapper, nil
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
