package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

var (
	lastPostDataMap OrderedMap
	mu              sync.RWMutex
)

func WriteRandomJsonFile(prefix string, body []byte) error {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(1000000)

	filename := filepath.Join("data", fmt.Sprintf("%s_%d_%06d.json", prefix, timestamp, random))

	if err := os.WriteFile(filename, body, 0644); err != nil {
		return err
	}

	return nil
}

func WriteRandomJsonFileIndented(prefix string, body []byte, jsonData any) error {
	indented, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return WriteRandomJsonFile(prefix, body)
	}
	return WriteRandomJsonFile(prefix, indented)
}

func GetKey(data any, path string) any {
	keys := strings.Split(path, ".")

	current := data
	for _, key := range keys {
		dataMap, ok := current.(map[string]interface{})
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

func WriteJsonResponse(body []byte) (int, error) {
	jsonCounter := 0

	// make sure the first byte is { (open curly brakets)
	if body[0] != '{' {
		return jsonCounter, nil
	}

	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		jsonCounter += 1
		return jsonCounter, WriteRandomJsonFileIndented("response", body, jsonData)
	}

	var lineData interface{}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] != '{' {
			continue
		}
		lineByte := []byte(line)
		if err := json.Unmarshal(lineByte, &lineData); err == nil {
			jsonCounter += 1
			WriteRandomJsonFileIndented("response", lineByte, lineData)
		}
	}

	return jsonCounter, nil

}

type OrderedMap struct {
	data  map[string]string
	order []string
}

func (om *OrderedMap) Set(key string, value string) {
	if _, exists := om.data[key]; !exists {
		om.order = append(om.order, key)
	}
	om.data[key] = value
}

func (om *OrderedMap) Get(key string) (string, bool) {
	data, exists := om.data[key]
	return data, exists
}

func (om *OrderedMap) Keys() []string {
	return om.order
}

func NewOrderedMap() OrderedMap {
	return OrderedMap{
		data:  map[string]string{},
		order: []string{},
	}
}

func GetPostDataMap(response playwright.Response) (OrderedMap, error) {
	om := NewOrderedMap()
	req := response.Request()
	data, err := req.PostData()
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

	ctx.OnResponse(func(response playwright.Response) {
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
			WriteJsonResponse(body)
			postDataMap, err := GetPostDataMap(response)
			if err != nil {
				fmt.Printf("Error GetPostDataMap(): %v\n", err)
				return
			}
			mu.Lock()
			lastPostDataMap = postDataMap
			mu.Unlock()
		}(response)
	})

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

	ForEachDetail(func(jsonData any) {
		productId := GetKey(jsonData, "ID")
		if productId == nil {
			return
		}
		description := GetKey(jsonData, "Description")
		if description != nil {
			return
		}
		// fmt.Printf("product %s does not have description\n", productId.(string))
		pages.GoToProduct(productId.(string))

		// sleep for 5 seconds
		time.Sleep(5 * time.Second)
	}, false)

	WaitingForInput()
}

func ForEachJsonInData(prefix string, process func(jsonData any), sortit bool) {
	// open and read all files in data folder that start with response and end in .json
	entries, err := os.ReadDir("data")
	if err != nil {
		log.Fatalf("error reading data directory: %v", err)
	}

	filePaths := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasPrefix(filename, prefix) || !strings.HasSuffix(filename, ".json") {
			continue
		}

		// Read and parse the JSON file
		filePath := filepath.Join("data", filename)

		filePaths = append(filePaths, filePath)
	}

	if sortit {
		// if sort is true sort filePaths in ascendant order
		sort.Strings(filePaths)
	} else {
		// else sort filePaths in random order
		rand.Shuffle(len(filePaths), func(i, j int) { filePaths[i], filePaths[j] = filePaths[j], filePaths[i] })
	}

	for _, filePath := range filePaths {

		// fmt.Printf("%s\n", filePath)

		body, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("error reading file %s: %v", filePath, err)
			continue
		}

		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			log.Printf("error parsing JSON from %s: %v", filePath, err)
			continue
		}

		process(jsonData)
	}

}

func ForEachResponse(process func(jsonData any), sortit bool) {
	ForEachJsonInData("response_", process, sortit)
}

func ForEachDetail(process func(jsonData any), sortit bool) {
	ForEachJsonInData("detail_", process, sortit)
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
	ForEachResponse(func(jsonData any) {
		productsA, _ := GetProductsFromSearch(jsonData)
		if hasAny := SaveProductsIfAny(productsA); hasAny == true {
			return
		}
		productsB, _ := GetProducFromData(jsonData)
		if hasAny := SaveProductsIfAny(productsB); hasAny == true {
			return
		}
		productsC, _ := GetProductDetails(jsonData)
		if hasAny := SaveProductsIfAny(productsC); hasAny == true {
			return
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
