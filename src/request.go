package main

import (
	"fmt"
	"strings"
	"time"

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

func RunRequestWithData(ctx ContextWrapperInterface, pwRequest playwright.Request, data string, headersSimple bool) (playwright.APIResponse, error) {
	url := pwRequest.URL()
	method := pwRequest.Method()

	headers, err := GetHeaders(pwRequest, headersSimple)
	if err != nil {
		return nil, fmt.Errorf("Error in GetHeaders: %w\n", err)
	}

	LogDebug0("RunRequest", "url", url, "method", method)
	LogDebug0("RunRequest request headers", "count", len(headers))
	for h, v := range headers {
		LogDebug0("RunRequest header", "key", h, "value", v)
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
	LogDebug0("RunRequest response headers", "count", len(responseHeaders))
	for h, v := range responseHeaders {
		LogDebug0("RunRequest response header", "key", h, "value", v)
	}

	return response, nil
}

func RunRequest(ctx ContextWrapperInterface, pwRequest playwright.Request, headersSimple bool) (playwright.APIResponse, error) {
	data, err := pwRequest.PostData()
	if err != nil {
		return nil, fmt.Errorf("Error in PostData: %w\n", err)
	}
	return RunRequestWithData(ctx, pwRequest, data, headersSimple)
}

func RunRequestDecompress(ctx ContextWrapperInterface, pwRequest playwright.Request, shouldSkipRequest ShouldProcess) (playwright.APIResponse, error) {

	var err error
	var newResponse playwright.APIResponse
	var requestCounter int

	LogInfo0("RunRequestDecompress starting", "initialRequests", requestCounter)
	ProcessData()

	ForEachDetail(func(filePath string, jsonData any) bool {

		description := GetKey(jsonData, "Description")
		if description != nil {
			return true
		}

		productId := GetKey(jsonData, "ID")
		if productId == nil {
			return true
		}

		shouldSkipRequest.postDataMap.SetJsonString("variables", "targetId", productId)

		newResponse, err = RunRequestDecompressOne(ctx, pwRequest, shouldSkipRequest)

		time.Sleep(1 * time.Second)

		requestCounter++

		if (requestCounter % 20) == 0 {
			time.Sleep(3 * time.Second)
			LogInfo0("RunRequestDecompress checkpoint", "requests", requestCounter)
			ProcessData()
		}

		return true

	}, false)

	time.Sleep(3 * time.Second)
	LogInfo0("RunRequestDecompress finished", "totalRequests", requestCounter)
	ProcessData()

	return newResponse, err
}

func RunRequestDecompressOne(ctx ContextWrapperInterface, pwRequest playwright.Request, shouldSkipRequest ShouldProcess) (playwright.APIResponse, error) {
	friendlyNameToProcess := "MarketplacePDPContainerQuery"

	if shouldSkipRequest.friendlyName != friendlyNameToProcess {
		return nil, nil
	}

	shouldSkipRequest.postDataMap.Print()
	postData, _ := GetPostData(shouldSkipRequest.postDataMap)

	response, err := RunRequestWithData(ctx, pwRequest, postData, false)
	if err != nil {
		return response, err
	}

	body, err := response.Body()
	if err != nil {
		return response, fmt.Errorf("Error response.Body(): %w\n", err)
	}

	bodyDecompressed, err := Decompress(body)
	if err != nil {
		return response, err
	}

	LogDebug0("Body decompressed", "size", len(bodyDecompressed))

	jsonDatas, err := ExtractJsonFromBody(bodyDecompressed)
	if err != nil {
		LogError0("Error ExtractJsonFromBody()", "error", err)
		return response, err
	}

	_, err = WriteJsonResponse(jsonDatas, friendlyNameToProcess)
	if err != nil {
		LogError0("Error WriteJsonResponse()", "error", err)
		return response, err
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

	var newBodyDecoded string
	decompressed, err := DecompressZstd(newBody)
	if err != nil {
		newBodyDecoded = string(newBody)
	} else {
		newBodyDecoded = string(decompressed)
	}

	if AreStringsEqual(bodyDecoded, newBodyDecoded) {
		LogInfo0("Response bodies are identical")
		return true, nil
	} else {
		bodyDecodedLen := len(bodyDecoded)
		newBodyDecodedLen := len(newBodyDecoded)
		LogWarn0("Response bodies differ", "originalLength", bodyDecodedLen, "newLength", newBodyDecodedLen)
		return false, nil
	}
}
