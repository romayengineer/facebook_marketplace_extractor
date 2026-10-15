package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

type ScrapperImpl struct {
	BlockImages             bool
	ScrollDownOnSearch      bool
	ScrollSeconds           int
	DelayMs                 int
	RateLimitWaitSeconds    int
	MaxRateLimitWaitSeconds int
	SearchKewords           string
	StartTimeToProcess      int64
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
		BlockImages:             blockImages,
		ScrollDownOnSearch:      ScrollDownOnSearch,
		ScrollSeconds:           flags.scrollSeconds,
		DelayMs:                 flags.delayMs,
		RateLimitWaitSeconds:    flags.rateLimitWaitSeconds,
		MaxRateLimitWaitSeconds: flags.maxRateLimitWaitSeconds,
		StartTimeToProcess:      0,
		SearchKewords:           flags.keywords,
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

	// Extract the listings from the search responses into data/__Details now, so the
	// search step materializes new products (with an empty Description) on its own.
	// Descriptions are still filled afterwards by pull_description.
	if _, err := ProcessData(0); err != nil {
		LogError0("SearchProducts", "Error in ProcessData", "error", err)
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

	// Materialize any as-yet-unprocessed search responses into data/__Details first,
	// so listings found by a recent `search` run are included in this fill pass.
	if _, err := ProcessData(0); err != nil {
		LogError0("PullDescriptions", "Error in ProcessData", "error", err)
	}

	if err := pages.MarketplaceHome(); err != nil {
		LogError0("PullDescriptions", "Error opening marketplace", "error", err)
		return
	}

	fmt.Println("Opening a listing to capture the PDP request template...")
	template := s.capturePDPTemplate(pages, page)
	if template == nil {
		LogError0("PullDescriptions", "did not capture the PDP request template", "hint", "Facebook is likely throttling this session (error 1675004); wait and retry later")
		return
	}

	existingPDPIds := LoadPDPProductIds()
	LogInfo0("PullDescriptions", "loaded existing PDP responses", "count", len(existingPDPIds))

	pulled := 0
	skipped := 0
	waitedSeconds := 0
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

		gotRateLimit, waited, err := s.fetchPDPWithBackoff(page, template, productIdStr)
		waitedSeconds += waited
		if err != nil {
			LogError0("PullDescriptions", "Error fetching PDP", "productId", productIdStr, "error", err)
			return true
		}
		if gotRateLimit {
			LogWarn0("PullDescriptions", "rate limit did not clear, stopping", "pulled", pulled, "waitedSec", waitedSeconds)
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
	if rateLimited {
		LogWarn0("PullDescriptions", "stopped early on rate limit; progress is saved, re-run later to continue (optionally -delay-ms=N)")
	}

	ProcessData(0)
}

// fetchPDPWithBackoff fetches a PDP and, if Facebook rate-limits the request (error
// code 1675004), waits and retries the same product until the limit clears or the
// per-run waiting budget (MaxRateLimitWaitSeconds) is exhausted. This makes a
// transient throttle pause the run instead of aborting it, while still saving all
// progress if the limit does not clear. Returns (rateLimited, waitedSeconds, err).
func (s *ScrapperImpl) fetchPDPWithBackoff(page PageWrapperInterface, template *PDPTemplate, productId string) (bool, int, error) {
	waited := 0
	wait := s.RateLimitWaitSeconds
	for {
		gotRateLimit, err := FetchPDP(page, template, productId)
		if err != nil {
			return false, waited, err
		}
		if !gotRateLimit {
			return false, waited, nil
		}

		if s.RateLimitWaitSeconds <= 0 {
			return true, waited, nil
		}
		if waited+wait > s.MaxRateLimitWaitSeconds {
			return true, waited, nil
		}

		// Exponential backoff (capped) so frequent retries do not keep the throttle
		// alive. A little jitter avoids retrying in lockstep.
		jitter := rand.Intn(wait/4 + 1)
		sleepFor := wait + jitter
		LogWarn0("PullDescriptions", "rate limited, waiting before retry", "productId", productId, "waitSec", sleepFor, "waitedSec", waited)
		time.Sleep(time.Duration(sleepFor) * time.Second)
		waited += sleepFor

		wait *= 2
		if wait > 600 {
			wait = 600
		}
	}
}

// capturePDPTemplate clicks listings until the browser issues a
// MarketplacePDPContainerQuery, which provides the reusable request template. If one
// listing does not produce the request (e.g. Facebook throttles it and serves the
// error page), it returns to the Marketplace home and tries the next listing.
func (s *ScrapperImpl) capturePDPTemplate(pages PagesInterface, page PageWrapperInterface) *PDPTemplate {
	const maxAttempts = 6
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			if err := pages.MarketplaceHome(); err != nil {
				LogWarn0("PullDescriptions", "could not return to marketplace", "attempt", attempt, "error", err)
			}
			time.Sleep(2 * time.Second)
		}

		if !pages.OpenProduct(attempt) {
			LogWarn0("PullDescriptions", "could not open a listing for the template", "attempt", attempt)
			continue
		}

		if template := waitForPDPTemplate(15); template != nil {
			return template
		}

		if throttled, _ := page.Evaluate(`document.body.innerText.includes("This page isn't available right now")`); throttled == true {
			LogWarn0("PullDescriptions", "Facebook is serving its error page for this session (likely throttled, code 1675004)")
		}
		LogWarn0("PullDescriptions", "no PDP request captured, trying another listing", "attempt", attempt)
	}
	return nil
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
