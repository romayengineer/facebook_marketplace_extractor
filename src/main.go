package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func GetKey(data any, path string) any {
	keys := strings.Split(path, ".")

	current := data
	for _, key := range keys {
		dataMap, ok := current.(map[string]any)
		if !ok {
			// err := fmt.Errorf("cannot access key %q: not a map", key)
			// fmt.Println(err)
			return nil
		}

		value, ok := dataMap[key]
		if !ok {
			// err := fmt.Errorf("key %q not found", key)
			// fmt.Println(err)
			return nil
		}
		current = value
	}

	return current
}

func ExtractJsonFromBody(body []byte) ([]any, error) {
	jsonDatas := []any{}
	// make sure the first byte is { (open curly brakets)
	if body[0] != '{' {
		return jsonDatas, nil
	}

	var jsonData any
	if err := json.Unmarshal(body, &jsonData); err == nil {
		jsonDatas = append(jsonDatas, jsonData)
		return jsonDatas, nil
	}

	var lineData any
	for line := range strings.SplitSeq(string(body), "\n") {
		if line[0] != '{' {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			jsonDatas = append(jsonDatas, lineData)
		}
	}

	return jsonDatas, nil

}

func WriteJsonResponse(jsonDatas []any, friendly_name string) (int, error) {
	jsonCounter := 0
	var err error
	for _, jsonData := range jsonDatas {
		if err = WriteRandomJsonFileIndented("response", friendly_name, jsonData); err != nil {
			return jsonCounter, err
		}
		jsonCounter += 1
	}

	return jsonCounter, nil
}

func Begin() (ContextWrapperInterface, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error NewConfig: %v", err)
	}

	playwrightWrapper, err := NewPlaywrightWrapper()
	if err != nil {
		return nil, fmt.Errorf("error NewPlaywrightWrapper: %v", err)
	}

	browser, err := playwrightWrapper.NewBrowser(false)
	if err != nil {
		return nil, fmt.Errorf("error NewBrowser: %v", err)
	}

	facebookScrapper := NewFacebookScrapper(browser)

	ctx, err := facebookScrapper.Login(config.UserCredentials)
	if err != nil {
		return nil, fmt.Errorf("error Login: %v", err)
	}

	// productExtractors := NewProductExtractors()

	SetContextEventHandlers(ctx)

	return ctx, nil
}

func SearchProducts() {
	ctx, err := Begin()
	if err != nil {
		log.Fatalf("error Begin: %v", err)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)
	pages.MarketpaceSearch("macbook")

	WaitingForInput()
}

func GetDetails() {
	ctx, err := Begin()
	if err != nil {
		log.Fatalf("error Begin: %v", err)
	}
	defer ctx.Close()

	page, _ := ctx.NewPage()
	pages, _ := NewPages(page)

	ForEachDetail(func(filePath string, jsonData any) {
		productId := GetKey(jsonData, "ID")
		if productId == nil {
			return
		}
		PriceAmount := GetKey(jsonData, "PriceAmount")
		if PriceAmount != nil {
			return
		}
		// fmt.Printf("product %s does not have description\n", productId.(string))
		pages.GoToProduct(productId.(string))

		// sleep for 5 seconds
		time.Sleep(3 * time.Second)
	}, false)

	WaitingForInput()
}

func SaveProductsIfAny(products []MarketplaceItemDetails) bool {
	if len(products) > 0 {
		for _, product := range products {
			store := NewProductFileStore(product.ID.(string))
			store.Save(product)
		}
		return true
	}
	return false
}

func ProcessData() {
	productExtractors := NewProductExtractors()
	ForEachResponse(func(filePath string, jsonData any) {
		for _, extractor := range productExtractors.extractors {
			product, _ := extractor.extractor(jsonData)
			if hasAny := SaveProductsIfAny(product); hasAny == true {
				return
			}
		}
		fmt.Printf("no product found deleting file: %s\n", filePath)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting file %s: %v\n", filePath, err)
		}
	}, true)
}

func main() {
	action := flag.String("action", "search", "Action to perform: search")
	flag.Parse()

	switch *action {
	case "search":
		SearchProducts()
	case "process_data":
		ProcessData()
	case "get_details":
		GetDetails()
	default:
		log.Fatalf("unknown action: %s", *action)
	}
}
