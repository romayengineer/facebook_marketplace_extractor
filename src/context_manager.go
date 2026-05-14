package main

import (
	"log"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type ContextEventHandlers struct {
	ctx ContextWrapperInterface
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
			newResponse, err := RunRequest(request, ceh.ctx)
			if err != nil {
				log.Printf("Error in RunRequest: %v", err)
				return
			}
			body, err := newResponse.Body()
			if err != nil {
				log.Printf("Error response.Body(): %v\n", err)
			}
			bodyEncoding := GuessEncoding(body)
			bodyDecoded, err := DecodeWithEncoding(body, bodyEncoding)
			if err != nil {
				log.Printf("Error DecodeWithEncoding(): %v\n", err)
				return
			}
			log.Printf("OnRequest body: %s\n", bodyDecoded[:100])
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
		// if friendlyName == "MarketplacePDPContainerQuery" {
		// 	newResponse, err := RunRequest(request, ctx)
		// 	if err != nil {
		// 		log.Printf("Error in RunRequest: %v", err)
		// 		return
		// 	}
		// 	CompareResponses(response, newResponse)
		// }
	}(response)
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
	if strings.HasSuffix(parsedURL.Path, ".jpg") {
		r.Abort()
		return
	}
	r.Continue()
}

func SetContextEventHandlers(ctx ContextWrapperInterface) {

	contextEventHandlers := ContextEventHandlers{
		ctx: ctx,
	}

	ctx.Route("**", contextEventHandlers.Route)

	// ctx.OnRequest(contextEventHandlers.OnRequest)

	ctx.OnResponse(contextEventHandlers.OnResponse)
}
