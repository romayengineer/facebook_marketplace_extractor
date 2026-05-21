package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/mozillazg/go-unidecode"
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

	// Add Unique index on facebook_id for fast lookups
	collection.Indexes = append(collection.Indexes, "CREATE UNIQUE INDEX IF NOT EXISTS `idx_facebook_id` ON `products` (facebook_id)")

	// Save the collection to database
	if err := db.app.Dao().SaveCollection(collection); err != nil {
		return fmt.Errorf("failed to create products collection: %v", err)
	}

	LogInfo0("ensureProductsCollection", "products collection created successfully")
	return nil
}

func SetProductFields(record *models.Record, product MarketplaceItemDetails) error {
	if record == nil {
		return fmt.Errorf("record is nil")
	}
	record.Set("facebook_id_long", toString(product.IDLong))
	record.Set("title", toStringClean(product.Title))
	record.Set("description", toStringClean(product.Description))
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

	return nil
}

func (db *PocketBaseDB) UpsertProduct(product MarketplaceItemDetails) (sql.Result, error) {

	facebookID := toString(product.ID)

	sql := `INSERT INTO products (
		facebook_id, facebook_id_long, title, description, category, url,
		price_amount, price_currency, creation_time, location_latitude,
		location_longitude, location_city_id, location_city_name1,
		location_city_name2, location_state_code, seller_id, seller_name,
		is_hidden, is_live, is_pending, is_sold
	) VALUES ({:facebook_id}, {:facebook_id_long}, {:title}, {:description}, {:category}, {:url},
		{:price_amount}, {:price_currency}, {:creation_time}, {:location_latitude},
		{:location_longitude}, {:location_city_id}, {:location_city_name1},
		{:location_city_name2}, {:location_state_code}, {:seller_id}, {:seller_name},
		{:is_hidden}, {:is_live}, {:is_pending}, {:is_sold})
	ON CONFLICT(facebook_id) DO UPDATE SET
		facebook_id_long = excluded.facebook_id_long,
		title = excluded.title,
		description = excluded.description,
		category = excluded.category,
		url = excluded.url,
		price_amount = excluded.price_amount,
		price_currency = excluded.price_currency,
		creation_time = excluded.creation_time,
		location_latitude = excluded.location_latitude,
		location_longitude = excluded.location_longitude,
		location_city_id = excluded.location_city_id,
		location_city_name1 = excluded.location_city_name1,
		location_city_name2 = excluded.location_city_name2,
		location_state_code = excluded.location_state_code,
		seller_id = excluded.seller_id,
		seller_name = excluded.seller_name,
		is_hidden = excluded.is_hidden,
		is_live = excluded.is_live,
		is_pending = excluded.is_pending,
		is_sold = excluded.is_sold`

	result, err := db.app.Dao().DB().NewQuery(sql).Bind(map[string]interface{}{
		"facebook_id":         facebookID,
		"facebook_id_long":    toString(product.IDLong),
		"title":               toStringClean(product.Title),
		"description":         toStringClean(product.Description),
		"category":            toString(product.Category),
		"url":                 toString(product.URL),
		"price_amount":        toFloat(product.PriceAmount),
		"price_currency":      toString(product.PriceCurrency),
		"creation_time":       toInt64(product.CreationTime),
		"location_latitude":   toFloat(product.LocationLatitud),
		"location_longitude":  toFloat(product.LocationLongitude),
		"location_city_id":    toString(product.LocationGeocodeCityID),
		"location_city_name1": toString(product.LocationGeocodeCityName1),
		"location_city_name2": toString(product.LocationGeocodeCityName2),
		"location_state_code": toString(product.LocationGeocodeStateCode),
		"seller_id":           toString(product.SellerID),
		"seller_name":         toString(product.SellerName),
		"is_hidden":           toBool(product.IsHidden),
		"is_live":             toBool(product.IsLive),
		"is_pending":          toBool(product.IsPending),
		"is_sold":             toBool(product.IsSold),
	}).Execute()

	if err != nil {
		LogError0("SaveProduct", "failed to upsert product", "facebook_id", facebookID, "error", err)
		return result, err
	}

	return result, nil
}

func (db *PocketBaseDB) GetProductID(product MarketplaceItemDetails) (string, error) {

	facebookID := toString(product.ID)

	var result struct {
		ID string
	}

	err := db.app.Dao().DB().NewQuery("SELECT id FROM products WHERE facebook_id = {:facebook_id} LIMIT 1").Bind(map[string]interface{}{
		"facebook_id": facebookID,
	}).One(&result)

	if err != nil {
		LogError0("SaveProduct", "failed to retrieve product id", "facebook_id", facebookID, "error", err)
		return "", err
	}

	return result.ID, nil
}

func (db *PocketBaseDB) SaveProduct(product MarketplaceItemDetails) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.UpsertProduct(product)
	productID, _ := db.GetProductID(product)

	LogDebug0("SaveProduct", "product saved", "productID", productID, "facebook_id", product.ID)
	return productID, nil
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

func toStringClean(v any) string {
	if v == nil {
		return ""
	}
	value := fmt.Sprintf("%v", v)
	// Convert UTF-8 to ASCII using unidecode
	value = unidecode.Unidecode(value)
	// Convert to lowercase
	value = strings.ToLower(value)
	return value
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

func ProcessDataInDB() (int64, error) {
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

		// description := GetKey(jsonData, "Description")
		// if description == nil {
		// 	return true
		// }

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
