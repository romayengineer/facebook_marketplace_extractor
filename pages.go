package main

import "net/url"

type Pages struct {
	Page          PageWrapperInterface
	InputUsername string
	InputPassword string
	ButtonLogIn   string
}

type PagesInterface interface {
	IsInHomePage() bool
	MarketpaceSearch(query string) error
}

func NewPages(page PageWrapperInterface) (PagesInterface, error) {
	pages := Pages{
		Page:          page,
		InputUsername: "input[name=email]",
		InputPassword: "input[name=pass]",
		ButtonLogIn:   "span:has-text('Log in')",
	}
	return &pages, nil
}

func (pl *Pages) IsInHomePage() bool {
	pl.Page.Locator("div[role=tablist]").WaitFor(10)
	items, _ := pl.Page.Locator("div[role=tablist]").All()
	return len(items) > 0
}

func (pl *Pages) MarketpaceSearch(query string) error {
	params := url.Values{}
	params.Add("query", query)
	baseUrl := "https://www.facebook.com/marketplace/category/search/?" + params.Encode()
	return pl.Page.Goto(baseUrl)
}
