package main

import (
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
	MarketpaceSearch(searchKeyworkds string, scrollDown bool) error
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
	pl.Page.Locator(pl.InputSearchBar).WaitFor(10)
	items, _ := pl.Page.Locator(pl.InputSearchBar).All()
	return len(items) > 0
}

func (pl *Pages) ScrollDownRelative() error {
	for {
		pl.Page.Evaluate("window.scrollBy(0, 300);")
		time.Sleep(2 * time.Second)
	}
}

func (pl *Pages) ScrollDown() error {
	maxScrollHeight := 0
	failCounter := 0
	maxFailCounter := 10
	for {
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
		pl.Page.Evaluate("window.scrollTo(0, document.body.scrollHeight)")
		time.Sleep(1 * time.Second)
	}
}

func (pl *Pages) MarketpaceSearch(searchKeyworkds string, scrollDown bool) error {
	// params := url.Values{}
	// params.Add("query", searchKeyworkds)
	// baseUrl := "https://www.facebook.com/marketplace/category/search/?" + params.Encode()
	baseUrl := "https://www.somoslamayor.com.ar/perfume"
	pl.Page.Goto(baseUrl)
	if scrollDown {
		return pl.ScrollDown()
	}
	return nil
}

func (pl *Pages) GoToProduct(id string) error {
	baseUrl := ProductIDTolink(id)
	LogInfo0("GoToProduct", "Goto", "url", baseUrl)
	return pl.Page.Goto(baseUrl)
}
