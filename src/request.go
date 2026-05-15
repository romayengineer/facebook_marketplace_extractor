package main

import (
	"encoding/json"
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

	log.Printf("RunRequest request headers count: %02d\n", len(headers))
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
	log.Printf("RunRequest response headers count: %02d\n", len(responseHeaders))
	for h, v := range responseHeaders {
		log.Printf("RunRequest response header: %s %s\n", h, v)
	}

	return response, nil
}

func DecompressJson(data []byte) (any, error) {
	var jsonData any

	if err := json.Unmarshal(data, &jsonData); err == nil {
		return jsonData, nil
	}

	for _, decompressor := range Decompressors {
		newData, err := decompressor(data)

		if err != nil {
			continue
		}

		if err := json.Unmarshal(newData, &jsonData); err == nil {
			return jsonData, nil
		}
	}

	return nil, fmt.Errorf("data is not json or compressed")
}

func RunRequestDecompress(ctx ContextWrapperInterface, pwRequest playwright.Request) (playwright.APIResponse, error) {
	response, err := RunRequest(ctx, pwRequest, false)
	if err != nil {
		return response, err
	}

	body, err := response.Body()
	if err != nil {
		return response, fmt.Errorf("Error response.Body(): %w\n", err)
	}

	_, err = DecompressJson(body)
	if err != nil {
		return nil, err
	}

	log.Printf("new response is JSON!")

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

	var newBodyDecoded string
	decompressed, err := DecompressZstd(newBody)
	if err != nil {
		newBodyDecoded = string(newBody)
	} else {
		newBodyDecoded = string(decompressed)
	}

	if AreStringsEqual(bodyDecoded, newBodyDecoded) {
		log.Printf("Response bodies same!\n\n\n\n")
		return true, nil
	} else {
		bodyDecodedLen := len(bodyDecoded)
		newBodyDecodedLen := len(newBodyDecoded)
		log.Printf("body: %s\n", bodyDecoded[:min(bodyDecodedLen, 300)])
		log.Printf("body length: %02d\n", bodyDecodedLen)
		log.Printf("newBody: %s\n", newBodyDecoded[:min(newBodyDecodedLen, 300)])
		log.Printf("newBody length: %02d\n", newBodyDecodedLen)
		log.Printf("Response bodies differ!\n\n\n\n")
		return false, nil
	}
}
