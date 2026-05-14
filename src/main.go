package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

var (
	lastPostDataMap        OrderedMap
	mu                     sync.RWMutex
	friendlyNamesToSkipSet = map[string]struct{}{
		"CometClassicHomeLeftRailContainerQuery":                    {},
		"CometFeedInlineComposerQuery":                              {},
		"CometHomeContactChannelsContainerQuery":                    {},
		"CometHomeContactCommunityChatsContainerQuery":              {},
		"CometHomeContactGroupsContainerQuery":                      {},
		"CometHomeContactsContainerQuery":                           {},
		"CometMarketplaceSetProductItemSeenStateMutation":           {},
		"CometMegaphoneRootQuery":                                   {},
		"CometModernHomeFeedQuery":                                  {},
		"CometNotificationsDropdownQuery":                           {},
		"CometRightSideHeaderCardsQuery":                            {},
		"CometSearchBootstrapKeywordsDataSourceQuery":               {},
		"FBScreenTimeLogger_syncMutation":                           {},
		"FBYRPTimeLimitsEnforcementQuery":                           {},
		"MAWVerifyThreadCutover_ContactCapabilities2Query":          {},
		"MarketplacePDPRightColumnAdsQuery":                         {},
		"OhaiWebClientMessengerConfigsQuery":                        {},
		"RTWebCallBlockSettingHooksQuery":                           {},
		"StoriesTrayRectangularRootQuery":                           {},
		"fetchMWChatVideoAutoplaySettingQuery":                      {},
		"useCIXLogMutation":                                         {},
		"useMWEncryptedBackupsFetchBackupIdsV2Query":                {},
		"usePseudoBlockedUserInterstitialF3Query":                   {},
		"useRainbowNativeSurveyDialogPlatformIntegrationPointQuery": {},
		// "CometMarketplaceSearchContentPaginationQuery":              {},
		// "MarketplaceCometBrowseFeedLightPaginationQuery":            {},
		// "MarketplacePDPC2CMediaViewerWithImagesQuery":               {},
		// "MarketplacePDPContainerQuery":                              {},
	}
)

func GetKey(data any, path string) any {
	keys := strings.Split(path, ".")

	current := data
	for _, key := range keys {
		dataMap, ok := current.(map[string]any)
		if !ok {
			// err := fmt.Errorf("cannot access key %q: not a map", key)
			// fmt.Println(err)
			return nil
		}

		value, ok := dataMap[key]
		if !ok {
			// err := fmt.Errorf("key %q not found", key)
			// fmt.Println(err)
			return nil
		}
		current = value
	}

	return current
}

func ExtractJsonFromBody(body []byte) ([]any, error) {
	jsonDatas := []any{}
	// make sure the first byte is { (open curly brakets)
	if body[0] != '{' {
		return jsonDatas, nil
	}

	var jsonData any
	if err := json.Unmarshal(body, &jsonData); err == nil {
		jsonDatas = append(jsonDatas, jsonData)
		return jsonDatas, nil
	}

	var lineData any
	for line := range strings.SplitSeq(string(body), "\n") {
		if line[0] != '{' {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			jsonDatas = append(jsonDatas, lineData)
		}
	}

	return jsonDatas, nil

}

func WriteJsonResponse(jsonDatas []any, friendly_name string) (int, error) {
	jsonCounter := 0
	var err error
	for _, jsonData := range jsonDatas {
		if err = WriteRandomJsonFileIndented("response", friendly_name, jsonData); err != nil {
			return jsonCounter, err
		}
		jsonCounter += 1
	}

	return jsonCounter, nil
}

type OrderedMap struct {
	data  map[string]string
	order []string
}

func (om *OrderedMap) Set(key string, value string) {
	if _, exists := om.data[key]; !exists {
		om.order = append(om.order, key)
	}
	om.data[key] = value
}

func (om *OrderedMap) Get(key string) (string, bool) {
	data, exists := om.data[key]
	return data, exists
}

func (om *OrderedMap) Keys() []string {
	return om.order
}

func (om *OrderedMap) Compare(om2 OrderedMap) {
	var v1 string
	var v2 string
	var exists bool
	equal := 0
	for _, k := range om.Keys() {
		if v2, exists = om2.data[k]; !exists {
			fmt.Printf("key changed %s\n", k)
		}
		if v1, exists = om.data[k]; !exists {
			fmt.Printf("key changed %s\n", k)
		}
		if v1 != v2 {
			fmt.Printf("key changed %s\n", k)
		} else {
			equal += 1
		}
	}
	fmt.Printf("keys equal: %02d\n\n\n", equal)
}

func NewOrderedMap() OrderedMap {
	return OrderedMap{
		data:  map[string]string{},
		order: []string{},
	}
}

func GetPostDataMap(request playwright.Request) (OrderedMap, error) {
	om := NewOrderedMap()
	data, err := request.PostData()
	if err != nil {
		return om, fmt.Errorf("error in req.PostData: %w\n", err)
	}
	decoded, err := url.QueryUnescape(data)
	if err != nil {
		return om, fmt.Errorf("error in url.QueryUnescape: %w\n", err)
	}
	for p := range strings.SplitSeq(decoded, "&") {
		key_value := strings.SplitN(p, "=", 2)
		if len(key_value) < 2 {
			return om, fmt.Errorf("key_value is not length 2: %s\n", p)
		}
		om.Set(key_value[0], key_value[1])
	}
	return om, nil
}

func Begin() (ContextWrapperInterface, error) {
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

	ctx.OnResponse(func(response playwright.Response) {
		go func(resp playwright.Response) {
			request := response.Request()
			url := request.URL()
			if url != "https://www.facebook.com/api/graphql/" {
				return
			}
			body, err := response.Body()
			if err != nil {
				fmt.Printf("Error response.Body(): %v\n", err)
				return
			}
			jsonDatas, err := ExtractJsonFromBody(body)
			if err != nil {
				fmt.Printf("Error ExtractJsonFromBody(): %v\n", err)
				return
			}
			// #TODO
			// for _, jsonData := range jsonDatas {
			// 	for _, extractor := range productExtractors.extractors {
			// 		valid := extractor.validator(jsonData)
			// 		if valid {

			// 		}
			// 	}
			// }
			mu.Lock()
			postDataMap, err := GetPostDataMap(request)
			if err != nil {
				fmt.Printf("Error GetPostDataMap(): %v\n", err)
				mu.Unlock()
				return
			} else {
				// postDataMap.Compare(lastPostDataMap)
				lastPostDataMap = postDataMap
			}
			mu.Unlock()
			var friendlyName string
			val, exists := postDataMap.Get("fb_api_req_friendly_name")
			if exists {
				friendlyName = val
			} else {
				friendlyName = "unknown"
			}
			if _, exists := friendlyNamesToSkipSet[friendlyName]; exists {
				return
			}
			_, err = WriteJsonResponse(jsonDatas, friendlyName)
			if err != nil {
				fmt.Printf("Error WriteJsonResponse(): %v\n", err)
			}
			// if friendlyName == "MarketplacePDPContainerQuery" {
			// 	newResponse, _ := RunRequest(request, ctx)
			// 	CompareResponses(response, newResponse)
			// }
		}(response)
	})

	return ctx, nil
}

func SearchProducts() {
	ctx, err := Begin()
	if err != nil {
		log.Fatalf("error Begin: %v", err)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()
}

func GetDetails() {
	ctx, err := Begin()
	if err != nil {
		log.Fatalf("error Begin: %v", err)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)

	ForEachDetail(func(filePath string, jsonData any) {
		productId := GetKey(jsonData, "ID")
		if productId == nil {
			return
		}
		description := GetKey(jsonData, "Description")
		if description != nil {
			return
		}
		// fmt.Printf("product %s does not have description\n", productId.(string))
		pages.GoToProduct(productId.(string))

		// sleep for 5 seconds
		time.Sleep(5 * time.Second)
	}, false)

	WaitingForInput()
}

func SaveProductsIfAny(products []MarketplaceItemDetails) bool {
	if len(products) > 0 {
		for _, product := range products {
			store := NewProductFileStore(product.ID.(string))
			store.Save(product)
		}
		return true
	}
	return false
}

func ProcessData() {
	productExtractors := NewProductExtractors()
	ForEachResponse(func(filePath string, jsonData any) {
		for _, extractor := range productExtractors.extractors {
			product, _ := extractor.extractor(jsonData)
			if hasAny := SaveProductsIfAny(product); hasAny == true {
				return
			}
		}
		fmt.Printf("no product found deleting file: %s\n", filePath)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting file %s: %v\n", filePath, err)
		}
	}, true)
}

func main() {
	action := flag.String("action", "search", "Action to perform: search")
	flag.Parse()

	switch *action {
	case "search":
		SearchProducts()
	case "process_data":
		ProcessData()
	case "get_details":
		GetDetails()
	default:
		log.Fatalf("unknown action: %s", *action)
	}
}
