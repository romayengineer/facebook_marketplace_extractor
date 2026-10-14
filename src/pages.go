package main

import (
	"net/url"
	"strings"
	"time"
)

type Pages struct {
	Page           PageWrapperInterface
	InputUsername  string
	InputPassword  string
	ButtonLogIn    string
	InputSearchBar string
}

type PagesInterface interface {
	IsInHomePage() bool
	WaitForHomePage(timeoutSeconds int) bool
	MarketpaceSearch(searchKeyworkds string, scrollDown bool, scrollSeconds int) error
	OpenFirstProduct() bool
	GoToProduct(id string) error
}

func NewPages(page PageWrapperInterface) (PagesInterface, error) {
	pages := Pages{
		Page:           page,
		InputUsername:  "input[name=email]",
		InputPassword:  "input[name=pass]",
		ButtonLogIn:    "span:has-text('Log in')",
		InputSearchBar: "input[placeholder='Buscar en Facebook']",
	}
	return &pages, nil
}

func (pl *Pages) IsInHomePage() bool {
	pl.Page.Locator(pl.InputSearchBar).WaitFor(30_000)
	return pl.isLoggedIn()
}

func (pl *Pages) isLoggedIn() bool {
	items, _ := pl.Page.Locator(pl.InputSearchBar).All()
	if len(items) > 0 {
		return true
	}

	urlAny, err := pl.Page.Evaluate("location.href")
	if err != nil {
		return false
	}
	currentURL, ok := urlAny.(string)
	if !ok {
		return false
	}
	if !strings.Contains(currentURL, "facebook.com") {
		return false
	}
	return !strings.Contains(currentURL, "/login") && !strings.Contains(currentURL, "checkpoint")
}

func (pl *Pages) WaitForHomePage(timeoutSeconds int) bool {
	for i := 0; i < timeoutSeconds; i++ {
		if pl.isLoggedIn() {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func (pl *Pages) ScrollDown(maxSeconds int) error {
	maxScrollHeight := 0
	failCounter := 0
	maxFailCounter := 10
	start := time.Now()
	for {
		if maxSeconds > 0 && time.Since(start) >= time.Duration(maxSeconds)*time.Second {
			LogInfo0("ScrollDown", "max scroll time reached", "seconds", maxSeconds, "maxScrollHeight", maxScrollHeight)
			return nil
		}
		scrollHeight, _ := pl.Page.Evaluate("document.body.scrollHeight")
		if intHeight, ok := scrollHeight.(int); ok {
			if intHeight > maxScrollHeight {
				failCounter = 0
				maxScrollHeight = intHeight
				LogDebug0("ScrollDown", "maxScrollHeight", "height", maxScrollHeight)
			} else {
				failCounter++
			}
		}
		if failCounter >= maxFailCounter {
			LogInfo0("ScrollDown", "maxScrollHeight did not increase", "failCounter", failCounter, "maxScrollHeight", maxScrollHeight)
			return nil
		}
		time.Sleep(1 * time.Second)
		pl.Page.Evaluate("window.scrollTo(0, document.body.scrollHeight)")
	}
}

func (pl *Pages) MarketpaceSearch(searchKeyworkds string, scrollDown bool, scrollSeconds int) error {
	params := url.Values{}
	params.Add("query", searchKeyworkds)
	baseUrl := "https://www.facebook.com/marketplace/category/search/?" + params.Encode()
	pl.Page.Goto(baseUrl)
	if scrollDown {
		return pl.ScrollDown(scrollSeconds)
	}
	return nil
}

func (pl *Pages) GoToProduct(id string) error {
	baseUrl := ProductIDTolink(id)
	LogInfo0("GoToProduct", "Goto", "url", baseUrl)
	return pl.Page.Goto(baseUrl)
}

// OpenFirstProduct clicks the first Marketplace listing on the search page.
//
// This is required to obtain `redacted_description`: Facebook only sends the
// MarketplacePDPContainerQuery request (which contains the description) when a
// listing is opened through a client-side click. A direct Goto does not reliably
// trigger it, so we click the link instead. See docs/redacted_description.md.
func (pl *Pages) OpenFirstProduct() bool {
	link := pl.Page.Locator(`a[href*="/marketplace/item/"]`).Nth(0)
	if err := link.Click(); err != nil {
		LogWarn0("OpenFirstProduct", "click failed, falling back to direct navigation", "error", err)
		return pl.openFirstProductByHref()
	}
	LogInfo0("OpenFirstProduct", "clicked first product link")
	return true
}

func (pl *Pages) openFirstProductByHref() bool {
	result, err := pl.Page.Evaluate(`(() => { const a = document.querySelector('a[href*="/marketplace/item/"]'); return a ? a.href : ""; })()`)
	if err != nil {
		LogError0("OpenFirstProduct", "Error evaluating selector", "error", err)
		return false
	}

	href, ok := result.(string)
	if !ok || href == "" {
		LogWarn0("OpenFirstProduct", "no marketplace item link found on page")
		return false
	}

	LogInfo0("OpenFirstProduct", "opening", "url", href)
	if err := pl.Page.Goto(href); err != nil {
		LogError0("OpenFirstProduct", "Error Goto", "url", href, "error", err)
		return false
	}
	return true
}
