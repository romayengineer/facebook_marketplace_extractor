package main

import (
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

var mu sync.RWMutex
var lastPostDataMap OrderedMap
var friendlyNamesToSkipSet = map[string]struct{}{
	"CometActorGatewayHandlerQuery":                                              {},
	"CometClassicHomeLeftRailContainerQuery":                                     {},
	"CometFeedInlineComposerQuery":                                               {},
	"CometHomeContactChannelsContainerQuery":                                     {},
	"CometHomeContactCommunityChatsContainerQuery":                               {},
	"CometHomeContactGroupsContainerQuery":                                       {},
	"CometHomeContactsContainerQuery":                                            {},
	"CometHomeCreateMenuContentQuery":                                            {},
	"CometHomeMegaMenuAllProductsQuery":                                          {},
	"CometHovercardQueryRendererQuery":                                           {},
	"CometLinkSharingInlineDisclosureCountQuery":                                 {},
	"CometMarketplaceSetProductItemSeenStateMutation":                            {},
	"CometMegaphoneRootQuery":                                                    {},
	"CometMessagingJewelDropdownEBUpsellContainerQuery":                          {},
	"CometMessagingJewelDropdownOnboardingUpsellQuery":                           {},
	"CometModernHomeFeedQuery":                                                   {},
	"CometNotificationsDropdownQuery":                                            {},
	"CometRightSideHeaderCardsQuery":                                             {},
	"CometSearchBootstrapKeywordsDataSourceQuery":                                {},
	"CometUnifiedShareSheetDialogQuery":                                          {},
	"FBScreenTimeLogger_syncMutation":                                            {},
	"FBYRPTimeLimitsEnforcementQuery":                                            {},
	"fetchMWChatVideoAutoplaySettingQuery":                                       {},
	"MarketplaceNotificationsUpdateSeenStateMutation":                            {},
	"MarketplacePDPRightColumnAdsQuery":                                          {},
	"MAWFetchXMAData_fetchXmaPreviewDataQuery":                                   {},
	"MAWVerifyThreadCutover_ContactCapabilities2Query":                           {},
	"OhaiWebClientMessengerConfigsQuery":                                         {},
	"RTWebCallBlockSettingHooksQuery":                                            {},
	"StoriesTrayRectangularRootQuery":                                            {},
	"UnifiedShareSheetMessengerSectionQuery":                                     {},
	"useCIXLogMutation":                                                          {},
	"useFeedComposerCometMentionsBootloadDataSourceQuery":                        {},
	"useFeedComposerCometMentionsBootloadDataSourceWithTaggingTransparencyQuery": {},
	"useMWEBDismissUpsellsOptOutEBHardblockReleaseMutation":                      {},
	"useMWEBOnboardingLogHardblockImpressionMutation":                            {},
	"useMWEncryptedBackupsFetchBackupIdsV2Query":                                 {},
	"usePseudoBlockedUserInterstitialF3Query":                                    {},
	"useRainbowNativeSurveyDialogPlatformIntegrationPointQuery":                  {},
	// "CometMarketplaceSearchContentPaginationQuery":                            {},
	// "MarketplaceCometBrowseFeedLightPaginationQuery":                          {},
	// "MarketplacePDPC2CMediaViewerWithImagesQuery":                             {},
	// "MarketplacePDPContainerQuery":                                            {},
}

type ContextEventHandlers struct {
	ctx              ContextWrapperInterface
	extensionsToSkip map[string]struct{}
}

func (ceh *ContextEventHandlers) OnRequest(request playwright.Request) {
	go func(request playwright.Request) {
		url := request.URL()
		if url != "https://www.facebook.com/api/graphql/" {
			return
		}
		// log.Printf("OnRequest request.URL(): %s\n", url)
		postDataMap, err := GetPostDataMap(request)
		if err != nil {
			return
		}
		var friendlyName string
		val, exists := postDataMap.Get("fb_api_req_friendly_name")
		if exists {
			friendlyName = val
		} else {
			friendlyName = "unknown"
		}
		if friendlyName == "MarketplacePDPContainerQuery" {
			newResponse, err := RunRequest(ceh.ctx, request, false)
			if err != nil {
				log.Printf("Error in RunRequest: %v", err)
				return
			}
			body, err := newResponse.Body()
			if err != nil {
				log.Printf("Error response.Body(): %v\n", err)
			}
			// bodyEncoding := GuessEncoding(body)
			// bodyDecoded, err := DecodeWithEncoding(body, bodyEncoding)
			// if err != nil {
			// 	log.Printf("Error DecodeWithEncoding(): %v\n", err)
			// 	return
			// }
			log.Printf("OnRequest body original: %s\n", body[:100])
			body, err = DecompressBrotli(body)
			if err != nil {
				log.Printf("Error DecompressBrotli(): %v\n", err)
				return
			}
			log.Printf("OnRequest body decompressed: %s\n", body[:100])
		}
	}(request)
}

func (ceh *ContextEventHandlers) OnResponse(response playwright.Response) {
	go func(resp playwright.Response) {
		request := response.Request()
		url := request.URL()
		if url != "https://www.facebook.com/api/graphql/" {
			return
		}
		body, err := response.Body()
		if err != nil {
			log.Printf("Error response.Body(): %v\n", err)
			return
		}
		jsonDatas, err := ExtractJsonFromBody(body)
		if err != nil {
			log.Printf("Error ExtractJsonFromBody(): %v\n", err)
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
			log.Printf("Error GetPostDataMap(): %v\n", err)
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
			log.Printf("Error WriteJsonResponse(): %v\n", err)
		}
		if friendlyName == "MarketplacePDPContainerQuery" {
			newResponse, err := RunRequest(ceh.ctx, request, false)
			if err != nil {
				log.Printf("Error in RunRequest: %v", err)
				return
			}
			CompareResponses(response, newResponse)
			newResponse, err = RunRequest(ceh.ctx, request, true)
			if err != nil {
				log.Printf("Error in RunRequest: %v", err)
				return
			}
			CompareResponses(response, newResponse)
		}
	}(response)
}

func GetExtension(path string) string {
	re := regexp.MustCompile(`\.[a-zA-Z0-9]+$`)
	match := re.FindString(path)
	return match
}

func ParseURL(urlString string) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Error parsing URL: %v\n", err)
		return
	}

	log.Printf("Protocol: %s\n", parsedURL.Scheme)
	log.Printf("Host: %s\n", parsedURL.Host)
	log.Printf("Hostname: %s\n", parsedURL.Hostname())
	log.Printf("Port: %s\n", parsedURL.Port())
	log.Printf("Path: %s\n", parsedURL.Path)
	log.Printf("Query: %s\n", parsedURL.RawQuery)
	log.Printf("Fragment: %s\n", parsedURL.Fragment)

	if parsedURL.RawQuery != "" {
		query := parsedURL.Query()
		for key, values := range query {
			log.Printf("Query param %s: %v\n", key, values)
		}
	}
}

func (ceh *ContextEventHandlers) Route(r playwright.Route) {
	request := r.Request()
	method := request.Method()
	if strings.ToLower(method) != "get" {
		r.Continue()
		return
	}
	urlString := request.URL()
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Error parsing URL: %v\n", err)
		r.Continue()
		return
	}

	extension := GetExtension(parsedURL.Path)
	if _, exists := ceh.extensionsToSkip[extension]; exists {
		r.Abort()
		return
	}
	r.Continue()
}

func SetContextEventHandlers(ctx ContextWrapperInterface) {

	contextEventHandlers := ContextEventHandlers{
		ctx: ctx,
		extensionsToSkip: map[string]struct{}{
			".jpg":  {},
			".webp": {},
			".mp3":  {},
			".mp4":  {},
			".svg":  {},
		},
	}

	ctx.Route("**", contextEventHandlers.Route)

	// ctx.OnRequest(contextEventHandlers.OnRequest)

	ctx.OnResponse(contextEventHandlers.OnResponse)
}
