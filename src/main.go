package main

import (
	"fmt"
	"os"
	"time"
)

type ScrapperImpl struct {
	BlockImages        bool
	ScrollDownOnSearch bool
	ScrollSeconds      int
	DelayMs            int
	SearchKewords      string
	StartTimeToProcess int64
}

func NewScrapper(flags Flags) ScrapperImpl {
	var blockImages bool
	var ScrollDownOnSearch bool
	switch flags.action {
	case "search":
		blockImages = true
		ScrollDownOnSearch = true
	case "pull_description":
		blockImages = true
	default:
		blockImages = false
		ScrollDownOnSearch = false
	}
	return ScrapperImpl{
		BlockImages:        blockImages,
		ScrollDownOnSearch: ScrollDownOnSearch,
		ScrollSeconds:      flags.scrollSeconds,
		DelayMs:            flags.delayMs,
		StartTimeToProcess: 0,
		SearchKewords:      flags.keywords,
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

	var ctx ContextWrapperInterface
	if config.UserDataDir != "" {
		ctx, err = playwrightWrapper.LaunchPersistentContext(config.UserDataDir, config.BrowserChannel, config.ProfileDirectory, false)
		if err != nil {
			return nil, fmt.Errorf("error LaunchPersistentContext: %v", err)
		}
		if err := EnsureLoggedIn(ctx); err != nil {
			return nil, fmt.Errorf("error EnsureLoggedIn: %v", err)
		}
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
	if err := pages.MarketpaceSearch(s.SearchKewords, s.ScrollDownOnSearch, s.ScrollSeconds); err != nil {
		LogError0("SearchProducts", "Error in MarketpaceSearch", "error", err)
	}

	WaitingForInput()
}

// PullDescriptions fetches the description (redacted_description) for every detail
// file that is missing one. It only needs to open a listing once, to capture the
// MarketplacePDPContainerQuery request template; the rest is replayed via ctx.Fetch.
// See docs/redacted_description.md.
func (s *ScrapperImpl) PullDescriptions() {
	ctx, err := s.Begin()
	if err != nil {
		LogError0("PullDescriptions", "Error in Begin", "error", err)
		os.Exit(1)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)

	if err := pages.MarketplaceHome(); err != nil {
		LogError0("PullDescriptions", "Error opening marketplace", "error", err)
		return
	}

	fmt.Println("Opening a listing to capture the PDP request template...")
	if !pages.OpenFirstProduct() {
		LogError0("PullDescriptions", "could not open a listing to capture the PDP template")
		return
	}

	template := waitForPDPTemplate(30)
	if template == nil {
		LogError0("PullDescriptions", "did not capture the PDP request template in time")
		return
	}

	existingPDPIds := LoadPDPProductIds()
	LogInfo0("PullDescriptions", "loaded existing PDP responses", "count", len(existingPDPIds))

	pulled := 0
	skipped := 0
	rateLimited := false
	ForEachDetail(func(filePath string, jsonData map[string]any) bool {
		description := GetKey(jsonData, "Description")
		if description != nil && description != "" {
			return true
		}

		productId := GetKey(jsonData, "ID")
		if productId == nil {
			return true
		}
		productIdStr, ok := productId.(string)
		if !ok {
			return true
		}

		// Already captured in a previous run: don't fetch (or write) again.
		if _, exists := existingPDPIds[productIdStr]; exists {
			skipped++
			return true
		}

		gotRateLimit, err := FetchPDP(page, template, productIdStr)
		if err != nil {
			LogError0("PullDescriptions", "Error fetching PDP", "productId", productIdStr, "error", err)
			return true
		}
		if gotRateLimit {
			LogWarn0("PullDescriptions", "rate limit reached, stopping", "pulled", pulled)
			rateLimited = true
			return false
		}

		existingPDPIds[productIdStr] = struct{}{}
		pulled++
		if pulled%50 == 0 {
			LogInfo0("PullDescriptions", "progress", "pulled", pulled)
		}
		if s.DelayMs > 0 {
			time.Sleep(time.Duration(s.DelayMs) * time.Millisecond)
		}
		return true
	}, false)

	LogInfo0("PullDescriptions", "finished", "pulled", pulled, "skipped", skipped, "rateLimited", rateLimited)

	ProcessData(0)
}

func waitForPDPTemplate(timeoutSeconds int) *PDPTemplate {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	for time.Now().Before(deadline) {
		if template, ok := GetPDPTemplate(); ok {
			return template
		}
		time.Sleep(250 * time.Millisecond)
	}
	return nil
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
		scrapper.PullDescriptions()
	case "get_details":
		scrapper.GetDetails()
	case "process_data":
		ProcessData(0)
	case "dedupe_pdp":
		DedupePDP()
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
