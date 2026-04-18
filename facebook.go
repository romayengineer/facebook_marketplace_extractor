package main

import (
	"fmt"
	"time"
)

type FacebookScrapperInterface interface {
	Login(userCredentials UserCredentials) error
}

type FacebookScrapper struct {
	Browser BrowserWrapperInterface
}

func NewFacebookScrapper(browser BrowserWrapperInterface) FacebookScrapperInterface {
	facebookScrapper := FacebookScrapper{
		Browser: browser,
	}
	return &facebookScrapper
}

func (fs *FacebookScrapper) Login(userCredentials UserCredentials) error {
	page, err := fs.Browser.NewPage()
	if err != nil {
		return fmt.Errorf("error NewPage: %v", err)
	}
	defer page.Close()

	url := "https://www.facebook.com"

	// Navigate to Facebook
	if err = page.Goto(url); err != nil {
		return fmt.Errorf("error Goto: %v", err)
	}

	fmt.Printf("Successfully opened %s in Chromium\n", url)

	emailInput := page.Locator("input[name=email]")
	emailInput.Fill(userCredentials.Username)

	passwordInput := page.Locator("input[name=pass]")
	passwordInput.Fill(userCredentials.Password)

	// Find and click the Log In button using getByRole
	loginButton := page.Locator("span:has-text('Log in')").GetByRole("button")
	err = loginButton.Click()
	if err != nil {
		fmt.Printf("error clicking Log In button: %v", err)
	}

	time.Sleep(10 * time.Minute)

	return nil
}
