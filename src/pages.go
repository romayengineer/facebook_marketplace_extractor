package main

import (
	"fmt"
	"log"
	"net/url"
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
	MarketpaceSearch(query string) error
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
				log.Printf("ScrollDown maxScrollHeight: %08d\n", maxScrollHeight)
			} else {
				failCounter++
			}
		}
		if failCounter >= maxFailCounter {
			log.Printf("ScrollDown maxScrollHeight did not increased for %02d times: %08d\n", maxFailCounter, maxScrollHeight)
			return nil
		}
		time.Sleep(1 * time.Second)
		pl.Page.Evaluate("window.scrollTo(0, document.body.scrollHeight)")
	}
}

func (pl *Pages) MarketpaceSearch(query string) error {
	params := url.Values{}
	params.Add("query", query)
	baseUrl := "https://www.facebook.com/marketplace/category/search/?" + params.Encode()
	pl.Page.Goto(baseUrl)
	return pl.ScrollDown()
}

func (pl *Pages) GoToProduct(id string) error {
	baseUrl := fmt.Sprintf("https://www.facebook.com/marketplace/item/%s", id)
	log.Printf("GoToProduct %s\n", baseUrl)
	return pl.Page.Goto(baseUrl)
}
