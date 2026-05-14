package main

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func RunRequest(pwRequest playwright.Request, ctx ContextWrapperInterface) (playwright.APIResponse, error) {
	url := pwRequest.URL()
	method := pwRequest.Method()
	data, _ := pwRequest.PostData()
	headers := pwRequest.Headers()

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

	return response, nil
}

func CompareResponses(response playwright.Response, newResponse playwright.APIResponse) (bool, error) {
	// if err != nil {
	// 	fmt.Printf("Error in RunRequest: %v\n", err)
	// 	return
	// }
	// defer newResponse.Dispose()

	body, err := response.Body()
	if err != nil {
		return false, fmt.Errorf("Error response.Body(): %w\n", err)
	}

	newBody, err := newResponse.Body()
	if err != nil {
		return false, fmt.Errorf("Error newResponse.Body(): %w\n", err)

	}

	if string(body) != string(newBody) {
		fmt.Printf("Response bodies differ!\n")
		fmt.Printf("newBody: %s\n", string(newBody))
		return false, nil
	} else {
		fmt.Printf("Response bodies same!\n")
		return true, nil
	}
}
