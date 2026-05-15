package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func GetPostDataMap(request playwright.Request) (OrderedMap, error) {
	om := NewOrderedMap()
	data, err := request.PostData()
	if err != nil {
		return om, fmt.Errorf("error in req.PostData: %w\n", err)
	}
	decoded, err := url.QueryUnescape(data)
	if err != nil {
		return om, fmt.Errorf("error in url.QueryUnescape: %w\n", err)
	}
	for p := range strings.SplitSeq(decoded, "&") {
		key_value := strings.SplitN(p, "=", 2)
		if len(key_value) < 2 {
			return om, fmt.Errorf("key_value is not length 2: %s\n", p)
		}
		om.Set(key_value[0], key_value[1])
	}
	return om, nil
}

func GetPostData(postData OrderedMap) (string, error) {
	postDataParts := []string{}
	for _, key := range postData.Keys() {
		value, exists := postData.Get(key)
		if !exists {
			continue
		}
		part := fmt.Sprintf("%s=%s", key, value)
		postDataParts = append(postDataParts, part)
	}
	postDataStr := strings.Join(postDataParts, "&")
	postDataUnscaped, err := url.QueryUnescape(postDataStr)
	return postDataUnscaped, err
}
