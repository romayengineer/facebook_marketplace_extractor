package main

import (
	"fmt"
	"strings"
)

type AttributeData struct {
	AttributeName string
	Label         string
	Value         string
}

type ItemLocation struct {
	Latitude  float64
	Longitude float64
}

type MarketplaceItemDetails struct {
	ID                       any
	TaxonomiPathJoined       any
	IDLong                   any
	Category                 any
	URL                      any
	Title                    any
	Description              any
	PriceAmount              any
	PriceCurrency            any
	AttributeData            any
	CreationTime             any
	LocationLatitud          any
	LocationLongitude        any
	LocationGeocodeCityID    any
	LocationGeocodeCityName1 any
	LocationGeocodeCityName2 any
	LocationGeocodeStateCode any
	SellerID                 any
	SellerName               any
	Photos                   any
	PhotoPrimary             any
	DeliveryTypes            any
	IsHidden                 any
	IsLive                   any
	IsPending                any
	IsSold                   any
}

type ProductExtractor struct {
	validator func(data any) bool
	extractor func(data any) ([]MarketplaceItemDetails, error)
}

type ProductExtractors struct {
	extractors []ProductExtractor
}

func NewProductExtractors() ProductExtractors {
	return ProductExtractors{
		extractors: []ProductExtractor{
			{
				validator: ProductDetailsValid,
				extractor: ProductDetailsGet,
			},
			{
				validator: ProductsFromSearchValid,
				extractor: ProductsFromSearchGet,
			},
			{
				validator: ProducFromDataValid,
				extractor: ProducFromDataGet,
			},
		},
	}
}

func ToMarketplaceItemDetails(data map[string]any) MarketplaceItemDetails {
	return MarketplaceItemDetails{
		ID:                       data["ID"],
		IDLong:                   data["IDLong"],
		Category:                 data["Category"],
		URL:                      data["URL"],
		Title:                    data["Title"],
		Description:              data["Description"],
		PriceAmount:              data["PriceAmount"],
		PriceCurrency:            data["PriceCurrency"],
		AttributeData:            data["AttributeData"],
		CreationTime:             data["CreationTime"],
		LocationLatitud:          data["LocationLatitud"],
		LocationLongitude:        data["LocationLongitude"],
		LocationGeocodeCityID:    data["LocationGeocodeCityID"],
		LocationGeocodeCityName1: data["LocationGeocodeCityName1"],
		LocationGeocodeCityName2: data["LocationGeocodeCityName2"],
		LocationGeocodeStateCode: data["LocationGeocodeStateCode"],
		SellerID:                 data["SellerID"],
		SellerName:               data["SellerName"],
		Photos:                   data["Photos"],
		PhotoPrimary:             data["PhotoPrimary"],
		DeliveryTypes:            data["DeliveryTypes"],
		IsHidden:                 data["IsHidden"],
		IsLive:                   data["IsLive"],
		IsPending:                data["IsPending"],
		IsSold:                   data["IsSold"],
	}
}

func ProductIDTolink(productId string) string {
	return fmt.Sprintf("https://www.facebook.com/marketplace/item/%s/", productId)
}

func IsErrorRateLimit(data any) bool {
	errors := GetKey(data, "errors")

	errorsList, ok := errors.([]any)
	if !ok {
		return false
	}

	for _, err := range errorsList {
		message := GetKey(err, "message")

		if message == nil {
			continue
		}

		if strings.ToLower(message.(string)) == "rate limit exceeded" {
			return true
		}

	}

	return false
}

func ProductDetailsValid(data any) bool {
	id := GetKey(data, "data.viewer.marketplace_product_details_page.target.id")
	return id != nil
}

func GetTaxonomiPathJoined(detail any) any {
	taxonomyPath := GetKey(detail, "marketplace_listing_renderable_target.seo_virtual_category.taxonomy_path")

	taxonomiNames := []string{}
	taxonomyPathList, ok := taxonomyPath.([]any)
	if ok {
		for _, taxonomi := range taxonomyPathList {
			taxonomiName := GetKey(taxonomi, "seo_info.seo_url")
			if taxonomiName == nil {
				continue
			}
			taxonomiNames = append(taxonomiNames, taxonomiName.(string))
		}
	}
	var taxonomiPathJoined any
	if len(taxonomiNames) == 0 {
		taxonomiPathJoined = nil
	} else {
		taxonomiPathJoined = strings.Join(taxonomiNames, ",")
	}
	return taxonomiPathJoined
}

func ProductDetailsGet(data any) ([]MarketplaceItemDetails, error) {
	detail := GetKey(data, "data.viewer.marketplace_product_details_page")
	if detail == nil {
		return nil, fmt.Errorf("detail not found")
	}

	productId := GetKey(detail, "target.id")
	if productId == nil {
		return nil, fmt.Errorf("detail does not have id")
	}

	var products []MarketplaceItemDetails

	productIdLong := GetKey(detail, "target.product_item.id")

	productUrl := GetKey(detail, "target.story.url")
	if productUrl == nil {
		productUrl = ProductIDTolink(productId.(string))
	}
	productTitle := GetKey(detail, "target.marketplace_listing_title")
	productDescription := GetKey(detail, "target.redacted_description.text")
	productPriceAmount := GetKey(detail, "target.listing_price.amount")
	productPriceCurrency := GetKey(detail, "target.listing_price.currency")
	productAttributeData := GetKey(detail, "target.attribute_data")
	productCreation := GetKey(detail, "target.creation_time")

	location := GetKey(detail, "marketplace_listing_renderable_target.location")
	latitud := GetKey(location, "latitude")
	longitude := GetKey(location, "longitude")
	cityName1 := GetKey(location, "reverse_geocode.city")
	cityName2 := GetKey(location, "reverse_geocode.city_page.display_name")
	cityID := GetKey(location, "reverse_geocode.city_page.id")
	stateCode := GetKey(location, "reverse_geocode.state")

	taxonomiPathJoined := GetTaxonomiPathJoined(detail)

	detailSellerId := GetKey(detail, "target.marketplace_listing_seller.id")
	detailSellerName := GetKey(detail, "target.marketplace_listing_seller.name")
	detailPhotos := GetKey(detail, "target.listing_photos")

	marketplaceItemDetails := MarketplaceItemDetails{
		ID:                       productId,
		TaxonomiPathJoined:       taxonomiPathJoined,
		IDLong:                   productIdLong,
		URL:                      productUrl,
		Title:                    productTitle,
		Description:              productDescription,
		PriceAmount:              productPriceAmount,
		PriceCurrency:            productPriceCurrency,
		AttributeData:            productAttributeData,
		CreationTime:             productCreation,
		LocationLatitud:          latitud,
		LocationLongitude:        longitude,
		LocationGeocodeCityID:    cityID,
		LocationGeocodeCityName1: cityName1,
		LocationGeocodeCityName2: cityName2,
		LocationGeocodeStateCode: stateCode,
		SellerID:                 detailSellerId,
		SellerName:               detailSellerName,
		Photos:                   detailPhotos,
	}

	products = append(products, marketplaceItemDetails)

	return products, nil
}

func ProductsFromSearchValid(data any) bool {
	edges := GetKey(data, "data.marketplace_search.feed_units.edges")
	if edges == nil {
		return false
	}

	edgesList, ok := edges.([]any)
	if !ok {
		return false
	}

	for _, edge := range edgesList {
		listing := GetKey(edge, "node.listing")
		if listing == nil {
			continue
		}

		productId := GetKey(listing, "id")
		if productId != nil {
			return true
		}
	}

	return false
}

func ProductsFromSearchGet(data any) ([]MarketplaceItemDetails, error) {
	edges := GetKey(data, "data.marketplace_search.feed_units.edges")
	if edges == nil {
		return nil, fmt.Errorf("no marketplace search found")
	}

	edgesList, ok := edges.([]any)
	if !ok {
		return nil, fmt.Errorf("edges is not a list")
	}

	var products []MarketplaceItemDetails

	for _, edge := range edgesList {
		listing := GetKey(edge, "node.listing")
		if listing == nil {
			continue
		}

		productId := GetKey(listing, "id")
		if productId == nil {
			continue
		}

		productUrl := ProductIDTolink(productId.(string))

		title := GetKey(listing, "marketplace_listing_title")
		price := GetKey(listing, "listing_price.amount")
		time := GetKey(listing, "if_gk_just_listed_tag_on_search_feed.creation_time")

		isHidden := GetKey(listing, "is_hidden")
		isLive := GetKey(listing, "is_live")
		isPending := GetKey(listing, "is_pending")
		isSold := GetKey(listing, "is_sold")

		location := GetKey(listing, "location")
		latitud := GetKey(location, "latitude")
		longitude := GetKey(location, "longitude")
		cityName1 := GetKey(location, "reverse_geocode.city")
		cityName2 := GetKey(location, "reverse_geocode.city_page.display_name")
		cityID := GetKey(location, "reverse_geocode.city_page.id")
		stateCode := GetKey(location, "reverse_geocode.state")

		sellerId := GetKey(listing, "marketplace_listing_seller.id")
		sellerName := GetKey(listing, "marketplace_listing_seller.name")

		photoPrimary := GetKey(listing, "primary_listing_photo")

		productDeliveryTypes := GetKey(listing, "delivery_types")

		marketplaceItemDetails := MarketplaceItemDetails{
			ID:                       productId,
			URL:                      productUrl,
			Title:                    title,
			CreationTime:             time,
			PriceAmount:              price,
			LocationLatitud:          latitud,
			LocationLongitude:        longitude,
			LocationGeocodeCityID:    cityID,
			LocationGeocodeCityName1: cityName1,
			LocationGeocodeCityName2: cityName2,
			LocationGeocodeStateCode: stateCode,
			SellerID:                 sellerId,
			SellerName:               sellerName,
			PhotoPrimary:             photoPrimary,
			DeliveryTypes:            productDeliveryTypes,
			IsHidden:                 isHidden,
			IsLive:                   isLive,
			IsPending:                isPending,
			IsSold:                   isSold,
		}

		products = append(products, marketplaceItemDetails)
	}

	return products, nil
}

func ProducFromDataValid(data any) bool {
	id := GetKey(data, "data.node.entity_id")
	return id != nil
}

func ProducFromDataGet(data any) ([]MarketplaceItemDetails, error) {
	node := GetKey(data, "data.node")
	if node == nil {
		return nil, fmt.Errorf("data node not found")
	}

	productId := GetKey(node, "entity_id")
	if productId == nil {
		return nil, fmt.Errorf("product does not have id")
	}

	var products []MarketplaceItemDetails

	productIdLong := GetKey(node, "data.product_item_id")

	productTitle := GetKey(node, "data.title")

	productCategory := GetKey(node, "data.upsell_type")

	// do not save this caterogy
	if productCategory != nil && productCategory == "CATEGORY_MISCELLANEOUS_UPSELL" {
		productCategory = nil
	}

	productUrl := ProductIDTolink(productId.(string))

	productPriceCurrency := GetKey(node, "data.price.currency")

	location := GetKey(node, "entity.location")
	cityName1 := GetKey(location, "reverse_geocode.city")
	cityName2 := GetKey(location, "reverse_geocode.city_page.display_name")
	cityID := GetKey(location, "reverse_geocode.city_page.id")
	stateCode := GetKey(location, "reverse_geocode.state")

	productCreation := GetKey(node, "listing.creation_time")

	marketplaceItemDetails := MarketplaceItemDetails{
		ID:                       productId,
		URL:                      productUrl,
		IDLong:                   productIdLong,
		Category:                 productCategory,
		Title:                    productTitle,
		PriceCurrency:            productPriceCurrency,
		LocationGeocodeCityID:    cityID,
		LocationGeocodeCityName1: cityName1,
		LocationGeocodeCityName2: cityName2,
		LocationGeocodeStateCode: stateCode,
		CreationTime:             productCreation,
	}

	products = append(products, marketplaceItemDetails)

	return products, nil
}
