package main

import (
	"bufio"
	"fmt"
	"os"
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

func WaitingForInput() {
	fmt.Printf("waiting for input: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}

func (fs *FacebookScrapper) Login(userCredentials UserCredentials) (ContextWrapperInterface, error) {

	savedSession := LoadSession()
	if savedSession != nil {
		fmt.Println("Loading existing session...")
		ctx, err := fs.Browser.NewContext(savedSession, true)
		if err == nil {
			page, err := ctx.NewPage()
			if err == nil {
				// Verify session is still valid
				err = page.Goto("https://www.facebook.com")
				if err == nil {
					pages, _ := NewPages(page)
					if pages.IsInHomePage() {
						fmt.Println("Session restored successfully")
						page.Close()
						return ctx, nil
					}
				}
				page.Close()
			}
			ctx.Close()
		}
		fmt.Printf("session expired, deletting")
		DeleteSession()
	}

	fmt.Printf("creating new session")

	// Create new context without saved session
	ctx, err := fs.Browser.NewContext(nil, true)
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
	if len(loginButtons) != 1 {
		return nil, fmt.Errorf("log in button must be 1")
	}
	loginButton := page.Locator("span:has-text('Log in')").Nth(0)
	err = loginButton.Click()
	if err != nil {
		return nil, fmt.Errorf("error clicking Log In button: %v", err)
	}

	WaitingForInput()

	pages, _ := NewPages(page)
	if pages.IsInHomePage() == false {
		return nil, fmt.Errorf("not in home page")
	}

	// Save the session for future use
	storageState, err := ctx.StorageState()
	if err != nil {
		return nil, fmt.Errorf("error: could not get session: %v\n", err)
	}

	err = SaveSession(storageState)
	if err != nil {
		return nil, fmt.Errorf("error: could not save session: %v\n", err)
	}

	fmt.Println("Session saved successfully")

	return ctx, nil
}
