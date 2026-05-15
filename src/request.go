package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func GetHeaders(request playwright.Request, simple bool) (map[string]string, error) {

	headers := map[string]string{}

	if simple {

		headers = request.Headers()

	} else {

		headersDirty, err := request.AllHeaders()
		if err != nil {
			return nil, err
		}

		for h, v := range headersDirty {
			if strings.HasPrefix(h, ":") {
				continue
			}
			headers[h] = v
		}
	}

	keys := len(headers)

	log.Printf("GetHeaders headers count: %02d\n", keys)

	return headers, nil
}

func RunRequest(ctx ContextWrapperInterface, pwRequest playwright.Request, headersSimple bool) (playwright.APIResponse, error) {
	url := pwRequest.URL()
	method := pwRequest.Method()
	data, err := pwRequest.PostData()
	if err != nil {
		return nil, fmt.Errorf("Error in PostData: %w\n", err)
	}

	headers, err := GetHeaders(pwRequest, headersSimple)
	if err != nil {
		return nil, fmt.Errorf("Error in GetHeaders: %w\n", err)
	}

	log.Printf("RunRequest url: %s\n", url)
	log.Printf("RunRequest method: %s\n", method)

	// log.Printf("RunRequest data: %s\n", data)

	for h, v := range headers {
		log.Printf("RunRequest request header: %s : %s\n", h, v)
	}

	response, err := ctx.Fetch(url,
		playwright.APIRequestContextFetchOptions{
			Method:  &method,
			Headers: headers,
			Data:    data,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Error executing apiRequest: %w\n", err)
	}

	responseHeaders := response.Headers()
	for h, v := range responseHeaders {
		log.Printf("RunRequest response header: %s %s\n", h, v)
	}

	return response, nil
}

func CompareResponses(response playwright.Response, newResponse playwright.APIResponse) (bool, error) {

	if newResponse == nil {
		return false, fmt.Errorf("Error in CompareResponses, newResponse is null")
	}

	body, err := response.Body()
	if err != nil {
		return false, fmt.Errorf("Error response.Body(): %w\n", err)
	}

	newBody, err := newResponse.Body()
	if err != nil {
		return false, fmt.Errorf("Error newResponse.Body(): %w\n", err)
	}

	bodyDecoded := string(body)
	newBodyDecoded := string(newBody)

	// newBody, err = DecompressBrotli(newBody)
	// if err != nil {
	// 	return false, fmt.Errorf("Error DecompressBrotli(): %w\n", err)
	// }

	// bodyEncoding := GuessEncoding(body)
	// bodyDecoded, err := DecodeWithEncoding(body, bodyEncoding)
	// if err != nil {
	// 	return false, err
	// }
	// newBodyEncoding := GuessEncoding(newBody)
	// newBodyDecoded, err := DecodeWithEncoding(newBody, newBodyEncoding)
	// if err != nil {
	// 	return false, err
	// }

	// log.Printf("Body encoding: %s, NewBody encoding: %s\n", bodyEncoding, newBodyEncoding)

	if bodyDecoded != newBodyDecoded {
		log.Printf("body: %s\n", bodyDecoded[:min(len(bodyDecoded), 300)])
		log.Printf("newBody: %s\n", newBodyDecoded[:min(len(newBodyDecoded), 300)])
		log.Printf("Response bodies differ!\n\n\n\n")
		return false, nil
	} else {
		log.Printf("Response bodies same!\n\n\n\n")
		return true, nil
	}
}
