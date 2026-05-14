package main

import (
	"fmt"
	"log"

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
		// 	newResponse, err := RunRequest(request, ctx)
		// 	if err != nil {
		// 		log.Printf("Error in RunRequest: %v", err)
		// 		return
		// 	}
		// 	CompareResponses(response, newResponse)
		// }
	}(response)
}

func SetContextEventHandlers(ctx ContextWrapperInterface) {

	ContextEventHandlers := ContextEventHandlers{
		ctx: ctx,
	}

	ctx.OnRequest(ContextEventHandlers.OnRequest)

	ctx.OnResponse(ContextEventHandlers.OnResponse)
}
