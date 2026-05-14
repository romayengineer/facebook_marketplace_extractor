package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/saintfish/chardet"
)

func GuessEncoding(data []byte) string {
	detector := chardet.NewTextDetector()
	results, _ := detector.DetectAll(data)
	for _, result := range results {
		return result.Charset
	}
	return "utf-8"
}

func GetHeaders(request playwright.Request) (map[string]string, error) {
	// headers := request.Headers()

	headersDirty, err := request.AllHeaders()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	for h, v := range headersDirty {
		if strings.HasPrefix(h, ":") {
			continue
		}
		headers[h] = v
	}

	return headers, nil
}

func RunRequest(pwRequest playwright.Request, ctx ContextWrapperInterface) (playwright.APIResponse, error) {
	url := pwRequest.URL()
	method := pwRequest.Method()
	data, err := pwRequest.PostData()
	if err != nil {
		return nil, err
	}

	headers, err := GetHeaders(pwRequest)
	if err != nil {
		return nil, fmt.Errorf("Error in GetHeaders: %w\n", err)
	}

	log.Printf("RunRequest url: %s\n", url)
	log.Printf("RunRequest method: %s\n", method)
	// log.Printf("RunRequest data: %s\n", data)
	for h, v := range headers {
		log.Printf("RunRequest header: %s : %s\n", h, v)
	}
	log.Printf("RunRequest log:\n")

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
	log.Printf("RunRequest end:\n")

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

	bodyEncoding := GuessEncoding(body)
	newBodyEncoding := GuessEncoding(newBody)

	log.Printf("Body encoding: %s, NewBody encoding: %s\n", bodyEncoding, newBodyEncoding)

	if string(body) != string(newBody) {
		log.Printf("Response bodies differ!\n")
		// fmt.Printf("body: %s\n", string(body))
		// fmt.Printf("newBody: %s\n", string(newBody))
		return false, nil
	} else {
		log.Printf("Response bodies same!\n")
		return true, nil
	}
}
