package main

import (
	"fmt"
	"os"
	"time"
)

type ScrapperImpl struct {
	BlockImages                 bool
	ScrollDownOnSearch          bool
	SearchKewords               string
	StartTimeToProcess          int64
	PullDescriptionWithKeywords string
}

func NewScrapper(flags Flags) ScrapperImpl {
	var blockImages bool
	var ScrollDownOnSearch bool
	switch flags.action {
	case "search":
		blockImages = false
		ScrollDownOnSearch = true
	default:
		blockImages = false
		ScrollDownOnSearch = false
	}
	return ScrapperImpl{
		BlockImages:                 blockImages,
		ScrollDownOnSearch:          ScrollDownOnSearch,
		StartTimeToProcess:          0,
		SearchKewords:               flags.keywords,
		PullDescriptionWithKeywords: flags.titleKeywords,
	}
}

func (s *ScrapperImpl) Begin() (ContextWrapperInterface, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error NewConfig: %v", err)
	}

	playwrightWrapper, err := NewPlaywrightWrapper()
	if err != nil {
		return nil, fmt.Errorf("error NewPlaywrightWrapper: %v", err)
	}

	browser, err := playwrightWrapper.NewBrowser(false)
	if err != nil {
		return nil, fmt.Errorf("error NewBrowser: %v", err)
	}

	facebookScrapper := NewFacebookScrapper(browser)

	ctx, err := facebookScrapper.Login(config.UserCredentials)
	if err != nil {
		return nil, fmt.Errorf("error Login: %v", err)
	}

	// productExtractors := NewProductExtractors()

	SetContextEventHandlers(ctx, s)

	return ctx, nil
}

func (s *ScrapperImpl) SearchProducts() {
	ctx, err := s.Begin()
	if err != nil {
		LogError0("SearchProducts", "Error in Begin", "error", err)
		os.Exit(1)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch(s.SearchKewords, s.ScrollDownOnSearch)

	WaitingForInput()
}

func (s *ScrapperImpl) GetDetails() {
	ctx, err := s.Begin()
	if err != nil {
		LogError0("GetDetails", "Error in Begin", "error", err)
		os.Exit(1)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)

	ForEachDetail(func(filePath string, jsonData map[string]any) bool {
		PriceAmount, _ := GetKey(jsonData, "PriceAmount")
		if PriceAmount != nil {
			return true
		}

		productId, _ := GetKey(jsonData, "ID")
		if productId == nil {
			return true
		}

		pages.GoToProduct(productId.(string))
		time.Sleep(3 * time.Second)

		return true
	}, false)

	WaitingForInput()
}

func main() {
	flags := NewFlags()
	LogInfo0("main", "flags", "action", flags.action)

	scrapper := NewScrapper(flags)

	switch flags.action {
	case "search":
		scrapper.SearchProducts()
	case "pull_description":
		scrapper.SearchProducts()
	case "get_details":
		scrapper.GetDetails()
	case "process_data":
		ProcessData(0)
	case "serve":
		Serve()
	case "save":
		_, err := ProcessDataInDB()
		if err != nil {
			LogFatal(err)
		}
	case "fill_empty":
		FillEmpty()
	default:
		LogError0("main", "Unknown action", "action", flags.action)
		os.Exit(1)
	}
}
