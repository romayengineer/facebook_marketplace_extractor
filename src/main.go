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
		blockImages = true
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

	fmt.Println("STEP: starting playwright wrapper")
	playwrightWrapper, err := NewPlaywrightWrapper()
	if err != nil {
		return nil, fmt.Errorf("error NewPlaywrightWrapper: %v", err)
	}
	fmt.Println("STEP: playwright wrapper ready")

	var ctx ContextWrapperInterface
	if config.UserDataDir != "" {
		fmt.Println("STEP: launching persistent context")
		ctx, err = playwrightWrapper.LaunchPersistentContext(config.UserDataDir, config.BrowserChannel, config.ProfileDirectory, false)
		if err != nil {
			return nil, fmt.Errorf("error LaunchPersistentContext: %v", err)
		}
		fmt.Println("STEP: persistent context launched")
		if err := EnsureLoggedIn(ctx); err != nil {
			return nil, fmt.Errorf("error EnsureLoggedIn: %v", err)
		}
		fmt.Println("STEP: logged in")
	} else {
		browser, err := playwrightWrapper.NewBrowser(false)
		if err != nil {
			return nil, fmt.Errorf("error NewBrowser: %v", err)
		}

		facebookScrapper := NewFacebookScrapper(browser)

		ctx, err = facebookScrapper.Login(config.UserCredentials)
		if err != nil {
			return nil, fmt.Errorf("error Login: %v", err)
		}
	}

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

	urlAny, _ := page.Evaluate("location.href")
	titleAny, _ := page.Evaluate("document.title")
	bodyAny, _ := page.Evaluate("document.body.innerText")
	fmt.Println("SEARCH URL:", urlAny)
	fmt.Println("SEARCH TITLE:", titleAny)
	if bodyStr, ok := bodyAny.(string); ok {
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500]
		}
		fmt.Println("SEARCH BODY:", bodyStr)
	}

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
		PriceAmount := GetKey(jsonData, "PriceAmount")
		if PriceAmount != nil {
			return true
		}

		productId := GetKey(jsonData, "ID")
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
