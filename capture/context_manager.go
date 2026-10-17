package capture

import (
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/mxschmitt/playwright-go"
	"github.com/romayengineer/facebook_marketplace_extractor/browser/driver"
	"github.com/romayengineer/facebook_marketplace_extractor/core/codec"
	"github.com/romayengineer/facebook_marketplace_extractor/core/jsonutil"
	"github.com/romayengineer/facebook_marketplace_extractor/core/logging"
	"github.com/romayengineer/facebook_marketplace_extractor/core/orderedmap"
	"github.com/romayengineer/facebook_marketplace_extractor/marketplace/model"
	"github.com/romayengineer/facebook_marketplace_extractor/store/filestore"
)

var mu sync.RWMutex
var lastPostDataMap orderedmap.OrderedMap
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
	ctx              driver.ContextWrapperInterface
	extensionsToSkip map[string]struct{}
}

type ShouldProcess struct {
	skip         bool
	postDataMap  orderedmap.OrderedMap
	friendlyName string
}

// PDPTemplate is the reusable request captured from a real MarketplacePDPContainerQuery.
// It is used to replay the query for other listings (with variables.targetId replaced).
type PDPTemplate struct {
	URL         string
	Method      string
	Headers     map[string]string
	PostDataMap orderedmap.OrderedMap
}

var (
	pdpTemplateMu sync.Mutex
	pdpTemplate   *PDPTemplate
)

func SetPDPTemplate(template PDPTemplate) {
	pdpTemplateMu.Lock()
	defer pdpTemplateMu.Unlock()
	if pdpTemplate == nil {
		pdpTemplate = &template
		logging.LogInfo0("SetPDPTemplate", "captured PDP request template", "url", template.URL, "headerKeys", mapKeys(template.Headers), "postKeys", template.PostDataMap.Keys())
	}
}

func GetPDPTemplate() (*PDPTemplate, bool) {
	pdpTemplateMu.Lock()
	defer pdpTemplateMu.Unlock()
	return pdpTemplate, pdpTemplate != nil
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
			logging.LogError0("OnRequest", "Error in RunRequest", "error", err)
			return
		}
		body, err := newResponse.Body()
		if err != nil {
			logging.LogError0("OnRequest", "Error newResponse.Body()", "error", err)
		}
		logging.LogDebug0("OnRequest", " body original", "size", len(body))
		body, err = codec.DecompressBrotli(body)
		if err != nil {
			logging.LogError0("OnRequest", "Error DecompressBrotli()", "error", err)
			return
		}
		logging.LogDebug0("OnRequest", " body decompressed", "size", len(body))
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
			logging.LogError0("OnResponse", "Error resp.Body()", "error", err)
			return
		}
		jsonDatas, err := jsonutil.ExtractJsonFromBody(body)
		if err != nil {
			logging.LogError0("OnResponse", "Error ExtractJsonFromBody()", "error", err)
			return
		}

		// MarketplacePDPContainerQuery responses carry the listing description
		// (redacted_description). A PDP request is only sent when a listing is opened
		// in the browser, so we capture this first request as a reusable template and
		// replay it for the other listings (see FetchPDP / PullDescriptions).
		// See docs/redacted_description.md.
		if shouldSkipRequest.friendlyName == "MarketplacePDPContainerQuery" {
			rateLimited := false
			for _, jsonData := range jsonDatas {
				if model.IsErrorRateLimit(jsonData) {
					rateLimited = true
					break
				}
			}

			if rateLimited {
				// Do not persist a rate-limit error as if it were a listing response:
				// it has no target.id and would pollute the PDP dir.
				logging.LogWarn0("OnResponse", "PDP response was rate limited, not saving")
			} else {
				// Write idempotently by targetId so the click response and the replayed
				// response for the same product collapse into one file.
				if targetId := GetPostDataTargetId(shouldSkipRequest.postDataMap); targetId != "" {
					if _, err := filestore.WriteJsonResponseById(jsonDatas, shouldSkipRequest.friendlyName, targetId); err != nil {
						logging.LogError0("OnResponse", "Error WriteJsonResponseById()", "error", err)
					}
				} else if _, err := filestore.WriteJsonResponse(jsonDatas, shouldSkipRequest.friendlyName); err != nil {
					logging.LogError0("OnResponse", "Error WriteJsonResponse()", "error", err)
				}

				// Merge the description into data/__Details immediately, so the listing
				// opened to capture the template is filled without waiting for ProcessData.
				filestore.SaveProductsFrom(jsonDatas)
			}

			if headers, err := GetHeaders(request, false); err == nil {
				SetPDPTemplate(PDPTemplate{
					URL:         request.URL(),
					Method:      request.Method(),
					Headers:     headers,
					PostDataMap: shouldSkipRequest.postDataMap.Clone(),
				})
			}
			return
		}

		if _, err = filestore.WriteJsonResponse(jsonDatas, shouldSkipRequest.friendlyName); err != nil {
			logging.LogError0("OnResponse", "Error WriteJsonResponse()", "error", err)
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
		logging.LogError0("ParseURL", "Error parsing URL", "error", err)
		return
	}

	logging.LogDebug0("ParseURL", "URL details", "scheme", parsedURL.Scheme, "host", parsedURL.Host, "port", parsedURL.Port(), "path", parsedURL.Path)

	if parsedURL.RawQuery != "" {
		query := parsedURL.Query()
		for key, values := range query {
			logging.LogDebug0("ParseURL", "Query param", "key", key, "values", values)
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
		logging.LogError0("Route", "Error parsing URL", "error", err)
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

func SetContextEventHandlers(ctx driver.ContextWrapperInterface, blockImages bool) {

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

	if blockImages {
		// if contextEventHandlers.Route is enabled then the response
		// from graphqh is invalid because there are missing headers
		// but it does not matter on search
		ctx.Route("**", contextEventHandlers.Route)
	}

	// ctx.OnRequest(contextEventHandlers.OnRequest)

	ctx.OnResponse(contextEventHandlers.OnResponse)
}
