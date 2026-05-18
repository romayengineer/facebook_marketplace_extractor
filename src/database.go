package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

type PocketBaseDB struct {
	app *pocketbase.PocketBase
	mu  sync.Mutex
}

var dbInstance *PocketBaseDB
var dbOnce sync.Once

func GetPocketBaseDB() (*PocketBaseDB, error) {
	var err error
	dbOnce.Do(func() {
		pb := pocketbase.New()
		dbInstance = &PocketBaseDB{
			app: pb,
		}
	})
	return dbInstance, err
}

func (db *PocketBaseDB) Start() error {
	if db.app == nil {
		return fmt.Errorf("pocketbase app not initialized")
	}

	// Bootstrap creates the database tables if they don't exist
	if err := db.app.Bootstrap(); err != nil {
		LogError0("Start", "bootstrap failed", "error", err)
		return fmt.Errorf("failed to bootstrap pocketbase: %v", err)
	}

	// Ensure products collection exists
	if err := db.ensureProductsCollection(); err != nil {
		LogError0("Start", "failed to ensure products collection", "error", err)
		return fmt.Errorf("failed to ensure products collection: %v", err)
	}

	LogInfo0("Start", "database initialized successfully")
	return nil
}

func (db *PocketBaseDB) ensureProductsCollection() error {
	// Try to find existing collection
	collection, err := db.app.Dao().FindCollectionByNameOrId("products")
	if err == nil {
		LogDebug0("ensureProductsCollection", "products collection already exists")
		return nil
	}

	LogInfo0("ensureProductsCollection", "creating products collection")

	// Create new collection
	collection = &models.Collection{
		Name: "products",
		Type: "base",
	}

	// Create schema with all fields
	s := schema.Schema{}

	fields := []struct {
		name string
		typ  string
	}{
		{"facebook_id", "text"},
		{"facebook_id_long", "text"},
		{"title", "text"},
		{"description", "text"},
		{"category", "text"},
		{"url", "url"},
		{"price_amount", "number"},
		{"price_currency", "text"},
		{"creation_time", "number"},
		{"location_latitude", "number"},
		{"location_longitude", "number"},
		{"location_city_id", "text"},
		{"location_city_name1", "text"},
		{"location_city_name2", "text"},
		{"location_state_code", "text"},
		{"seller_id", "text"},
		{"seller_name", "text"},
		{"is_hidden", "bool"},
		{"is_live", "bool"},
		{"is_pending", "bool"},
		{"is_sold", "bool"},
	}

	for _, f := range fields {
		s.AddField(&schema.SchemaField{
			Name: f.name,
			Type: f.typ,
		})
	}

	collection.Schema = s

	// Save the collection to database
	if err := db.app.Dao().SaveCollection(collection); err != nil {
		return fmt.Errorf("failed to create products collection: %v", err)
	}

	LogInfo0("ensureProductsCollection", "products collection created successfully")
	return nil
}

func (db *PocketBaseDB) SaveProduct(product MarketplaceItemDetails) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	collection, err := db.app.Dao().FindCollectionByNameOrId("products")
	if err != nil {
		LogError0("SaveProduct", "collection not found", "error", err)
		return "", fmt.Errorf("products collection not found: %v", err)
	}

	facebookID := toString(product.ID)
	var record *models.Record

	// Try to find existing record by facebook_id
	existingRecord, err := db.app.Dao().FindFirstRecordByData(collection.Id, "facebook_id", facebookID)
	if err == nil && existingRecord != nil {
		// Update existing record
		record = existingRecord
		LogDebug0("SaveProduct", "updating existing product", "facebook_id", facebookID)
	} else {
		// Create new record
		record = models.NewRecord(collection)
		record.Set("facebook_id", facebookID)
		LogDebug0("SaveProduct", "creating new product", "facebook_id", facebookID)
	}

	// Set all fields
	record.Set("facebook_id_long", toString(product.IDLong))
	record.Set("title", toString(product.Title))
	record.Set("description", toString(product.Description))
	record.Set("category", toString(product.Category))
	record.Set("url", toString(product.URL))
	record.Set("price_amount", toFloat(product.PriceAmount))
	record.Set("price_currency", toString(product.PriceCurrency))
	record.Set("creation_time", toInt64(product.CreationTime))
	record.Set("location_latitude", toFloat(product.LocationLatitud))
	record.Set("location_longitude", toFloat(product.LocationLongitude))
	record.Set("location_city_id", toString(product.LocationGeocodeCityID))
	record.Set("location_city_name1", toString(product.LocationGeocodeCityName1))
	record.Set("location_city_name2", toString(product.LocationGeocodeCityName2))
	record.Set("location_state_code", toString(product.LocationGeocodeStateCode))
	record.Set("seller_id", toString(product.SellerID))
	record.Set("seller_name", toString(product.SellerName))
	record.Set("is_hidden", toBool(product.IsHidden))
	record.Set("is_live", toBool(product.IsLive))
	record.Set("is_pending", toBool(product.IsPending))
	record.Set("is_sold", toBool(product.IsSold))

	if err := db.app.Dao().SaveRecord(record); err != nil {
		LogError0("SaveProduct", "failed to save product", "facebook_id", facebookID, "error", err)
		return "", err
	}

	LogDebug0("SaveProduct", "product saved", "id", record.Id, "facebook_id", facebookID)
	return record.Id, nil
}

func (db *PocketBaseDB) SaveProducts(products []MarketplaceItemDetails) (int, error) {
	count := 0
	for _, product := range products {
		if _, err := db.SaveProduct(product); err != nil {
			LogError0("SaveProducts", "error saving product", "facebook_id", product.ID, "error", err)
			continue
		}
		count++
	}
	LogInfo0("SaveProducts", "batch save completed", "count", count, "total", len(products))
	return count, nil
}

func (db *PocketBaseDB) ProductExists(facebookID string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	collection, err := db.app.Dao().FindCollectionByNameOrId("products")
	if err != nil {
		return false, err
	}

	record, err := db.app.Dao().FindFirstRecordByData(collection.Id, "facebook_id", toString(facebookID))
	if err != nil {
		return false, nil
	}

	return record != nil, nil
}

func (db *PocketBaseDB) GetApp() *pocketbase.PocketBase {
	return db.app
}

func (db *PocketBaseDB) Close() error {
	return nil
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		num, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			LogFatal("cannot convert to float")
		}
		return num
	default:
		LogFatal("cannot convert to float")
	}
	return 0
}

func toInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		num, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			LogFatal("cannot convert to int")
		}
		return num
	default:
		LogFatal("cannot convert to int")
	}
	return 0
}

func toBool(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	default:
		return false
	}
}

func Serve() (*PocketBaseDB, error) {

	// need to override the flags because pocketbase reads them
	os.Args = os.Args[:1]
	os.Args = append(os.Args, "serve")

	dbInstance, err := GetPocketBaseDB()
	if err != nil {
		return dbInstance, err
	}
	err = dbInstance.app.Start()
	if err != nil {
		return dbInstance, err
	}
	return dbInstance, err
}

func ProcessDataInDB(startAtTimestamp int64) (int64, error) {
	var filesProcessedCounter int64

	dbInstance, err := GetPocketBaseDB()
	if err != nil {
		return filesProcessedCounter, err
	}
	err = dbInstance.Start()
	if err != nil {
		return filesProcessedCounter, err
	}

	ForEachDetail(func(filePath string, jsonData map[string]any) bool {

		description := GetKey(jsonData, "Description")
		if description == nil {
			return true
		}

		product := ToMarketplaceItemDetails(jsonData)
		productIdDb, err := dbInstance.SaveProduct(product)

		if err != nil {
			LogError0("ProcessDataInDB", "Error in dbInstance.SaveProduct", err)
			return false
		}

		LogInfo0("ProcessDataInDB", "saved product in db", "productIdDb", productIdDb, "product.ID", product.ID)

		filesProcessedCounter += 1

		return true

	}, true)

	LogInfo0("ProcessDataInDB", "all files processed", "filesProcessedCounter", filesProcessedCounter)

	return filesProcessedCounter, nil
}
