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

type ShouldProcess struct {
	skip         bool
	postDataMap  OrderedMap
	friendlyName string
}

func ShouldSkipRequest(request playwright.Request) ShouldProcess {
	url := request.URL()
	if url != "https://www.facebook.com/api/graphql/" {
		return ShouldProcess{skip: true}
	}
	postDataMap, err := GetPostDataMap(request)
	if err != nil {
		return ShouldProcess{skip: true, postDataMap: postDataMap}
	}
	mu.Lock()
	lastPostDataMap = postDataMap
	mu.Unlock()
	friendlyName := postDataMap.GetDefault("fb_api_req_friendly_name", "unknown")
	if _, exists := friendlyNamesToSkipSet[friendlyName]; exists {
		return ShouldProcess{skip: true, postDataMap: postDataMap, friendlyName: friendlyName}
	}
	return ShouldProcess{skip: false, postDataMap: postDataMap, friendlyName: friendlyName}
}

func (ceh *ContextEventHandlers) OnRequest(request playwright.Request) {
	go func(req playwright.Request) {
		shouldSkipRequest := ShouldSkipRequest(req)
		if shouldSkipRequest.skip {
			return
		}
		newResponse, err := RunRequest(ceh.ctx, req, false)
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
	}(request)
}

func (ceh *ContextEventHandlers) OnResponse(response playwright.Response) {
	go func(resp playwright.Response) {
		request := resp.Request()
		shouldSkipRequest := ShouldSkipRequest(request)
		if shouldSkipRequest.skip {
			return
		}
		body, err := resp.Body()
		if err != nil {
			log.Printf("Error resp.Body(): %v\n", err)
			return
		}
		jsonDatas, err := ExtractJsonFromBody(body)
		if err != nil {
			log.Printf("Error ExtractJsonFromBody(): %v\n", err)
			return
		}
		_, err = WriteJsonResponse(jsonDatas, shouldSkipRequest.friendlyName)
		if err != nil {
			log.Printf("Error WriteJsonResponse(): %v\n", err)
		}
		if shouldSkipRequest.friendlyName != "MarketplacePDPContainerQuery" {
			return
		}
		newResponse, err := RunRequestDecompress(ceh.ctx, request)
		if err != nil {
			log.Printf("Error in RunRequest: %v", err)
			return
		}
		CompareResponses(resp, newResponse)
		// newResponse, err = RunRequest(ceh.ctx, request, true)
		// if err != nil {
		// 	log.Printf("Error in RunRequest: %v", err)
		// 	return
		// }
		// CompareResponses(resp, newResponse)
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

	// if contextEventHandlers.Route is enabled then the response
	// from graphqh is invalid because there are missing headers
	// ctx.Route("**", contextEventHandlers.Route)

	// ctx.OnRequest(contextEventHandlers.OnRequest)

	ctx.OnResponse(contextEventHandlers.OnResponse)
}
