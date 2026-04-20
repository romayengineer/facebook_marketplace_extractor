package main

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type LocatorWrapperInterface interface {
	Fill(value string) error
	Click() error
	Locator(selector string) LocatorWrapperInterface
	WaitFor(timeout float64) error
	GetByRole(role playwright.AriaRole) LocatorWrapperInterface
	Nth(index int) LocatorWrapperInterface
	All() ([]LocatorWrapperInterface, error)
}

type ContextWrapperInterface interface {
	NewPage() (PageWrapperInterface, error)
	Close() error
	StorageState() (*playwright.StorageState, error)
}

type PageWrapperInterface interface {
	Goto(url string) error
	Locator(selector string) LocatorWrapperInterface
	Close() error
	GetByRole(role playwright.AriaRole) LocatorWrapperInterface
}

type BrowserWrapperInterface interface {
	NewContext(storageState *playwright.StorageState, isMobule bool) (ContextWrapperInterface, error)
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

type ContextWrapper struct {
	Context playwright.BrowserContext
}

type PageWrapper struct {
	Page playwright.Page
}

type BrowserWrapper struct {
	Browser    playwright.Browser
	Playwright *playwright.Playwright
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

func (lw *LocatorWrapper) WaitFor(timeout float64) error {
	err := lw.locator.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: &timeout,
	})
	return err
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

func (lw *LocatorWrapper) Nth(index int) LocatorWrapperInterface {
	locator := lw.locator.Nth(index)
	locatorWrapper := NewLocatorWrapper(locator)
	return locatorWrapper
}

func (lw *LocatorWrapper) All() ([]LocatorWrapperInterface, error) {
	locators, err := lw.locator.All()
	var locatorsWrapper []LocatorWrapperInterface
	for _, locator := range locators {
		locatorsWrapper = append(locatorsWrapper, NewLocatorWrapper(locator))
	}
	return locatorsWrapper, err
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

func NewBrowserWrapper(browser playwright.Browser, pw *playwright.Playwright) (BrowserWrapperInterface, error) {
	browserWrapper := BrowserWrapper{
		Browser:    browser,
		Playwright: pw,
	}
	return &browserWrapper, nil
}

func NewContextWrapper(context playwright.BrowserContext) ContextWrapperInterface {
	contextWrapper := ContextWrapper{Context: context}
	return &contextWrapper
}

func (ww *BrowserWrapper) NewContext(storageState *playwright.StorageState, isMobile bool) (ContextWrapperInterface, error) {
	opts := playwright.BrowserNewContextOptions{}
	if storageState != nil {
		opts.StorageState = storageState.ToOptionalStorageState()
	}
	if isMobile == true && ww.Playwright != nil {
		device := ww.Playwright.Devices["Pixel 5"]
		opts.Viewport = device.Viewport
		opts.UserAgent = playwright.String(device.UserAgent)
		opts.DeviceScaleFactor = playwright.Float(device.DeviceScaleFactor)
		opts.IsMobile = playwright.Bool(device.IsMobile)
		opts.HasTouch = playwright.Bool(device.HasTouch)
	}
	context, err := ww.Browser.NewContext(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not create context: %v", err)
	}
	return NewContextWrapper(context), nil
}

func (ww *BrowserWrapper) Close() error {
	return ww.Browser.Close()
}

func (cw *ContextWrapper) NewPage() (PageWrapperInterface, error) {
	page, err := cw.Context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("Could not create page: %v", err)
	}
	return NewPageWrapper(page), nil
}

func (cw *ContextWrapper) Close() error {
	return cw.Context.Close()
}

func (cw *ContextWrapper) StorageState() (*playwright.StorageState, error) {
	storageState, err := cw.Context.StorageState()
	if err != nil {
		return nil, fmt.Errorf("Could not get storage state: %v", err)
	}
	return storageState, nil
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
	browserWrapper, err := NewBrowserWrapper(browser, pw.Playwright)
	if err != nil {
		return browserWrapper, fmt.Errorf("Could not create BrowserWrapper: %v", err)
	}
	return browserWrapper, nil
}
