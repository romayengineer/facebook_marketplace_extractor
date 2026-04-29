package main

import (
	"fmt"
	"net/url"
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

func (pl *Pages) MarketpaceSearch(query string) error {
	params := url.Values{}
	params.Add("query", query)
	baseUrl := "https://www.facebook.com/marketplace/category/search/?" + params.Encode()
	return pl.Page.Goto(baseUrl)
}

func (pl *Pages) GoToProduct(id string) error {
	baseUrl := fmt.Sprintf("https://www.facebook.com/marketplace/item/%s", id)
	return pl.Page.Goto(baseUrl)
}
