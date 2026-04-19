package main

import (
	"fmt"
	"time"
)

type FacebookScrapperInterface interface {
	Login(userCredentials UserCredentials) (ContextWrapperInterface, error)
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

func (fs *FacebookScrapper) Login(userCredentials UserCredentials) (ContextWrapperInterface, error) {
	// Try to load existing session
	savedSession := LoadSession()
	if savedSession != nil {
		fmt.Println("Loading existing session...")
		ctx, err := fs.Browser.NewContext(savedSession)
		if err == nil {
			page, err := ctx.NewPage()
			if err == nil {
				// Verify session is still valid
				err = page.Goto("https://www.facebook.com")
				time.Sleep(4 * time.Minute)
				page.Close()
				if err == nil {
					fmt.Println("Session restored successfully")
					return ctx, nil
				}
			}
			ctx.Close()
		}
		// If session is invalid, delete it and continue with login
		DeleteSession()
	}

	// Create new context without saved session
	ctx, err := fs.Browser.NewContext(nil)
	if err != nil {
		return nil, fmt.Errorf("error NewContext: %v", err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("error NewPage: %v", err)
	}
	defer page.Close()

	url := "https://www.facebook.com"

	// Navigate to Facebook
	if err = page.Goto(url); err != nil {
		return nil, fmt.Errorf("error Goto: %v", err)
	}

	fmt.Printf("Successfully opened %s in Chromium\n", url)

	emailInput := page.Locator("input[name=email]")
	emailInput.Fill(userCredentials.Username)

	passwordInput := page.Locator("input[name=pass]")
	passwordInput.Fill(userCredentials.Password)

	// Find and click the Log In button using getByRole
	loginButtons, _ := page.Locator("span:has-text('Log in')").All()
	if len(loginButtons) != 4 {
		return nil, fmt.Errorf("something changed")
	}
	loginButton := page.Locator("span:has-text('Log in')").Nth(1)
	err = loginButton.Click()
	if err != nil {
		return nil, fmt.Errorf("error clicking Log In button: %v", err)
	}

	// Wait for navigation after login
	time.Sleep(4 * time.Minute)

	// Save the session for future use
	storageState, err := ctx.StorageState()
	if err != nil {
		fmt.Printf("warning: could not save session: %v\n", err)
	} else {
		err = SaveSession(storageState)
		if err != nil {
			fmt.Printf("warning: could not persist session: %v\n", err)
		} else {
			fmt.Println("Session saved successfully")
		}
	}

	return ctx, nil
}
