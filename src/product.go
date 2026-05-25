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

type Location struct {
	LocationLatitud          any
	LocationLongitude        any
	LocationGeocodeCityID    any
	LocationGeocodeCityName1 any
	LocationGeocodeCityName2 any
	LocationGeocodeStateCode any
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
		TaxonomiPathJoined:       data["TaxonomiPathJoined"],
		IDLong:                   data["IDLong"],
		Category:                 data["Category"],
		URL:                      data["URL"],
		Title:                    data["Title"],
		Description:              data["Description"],
		PriceAmount:              data["PriceAmount"],
		PriceCurrency:            data["PriceCurrency"],
		AttributeData:            data["AttributeData"],
		CreationTime:             data["CreationTime"],
		SellerID:                 data["SellerID"],
		SellerName:               data["SellerName"],
		Photos:                   data["Photos"],
		PhotoPrimary:             data["PhotoPrimary"],
		DeliveryTypes:            data["DeliveryTypes"],
		IsHidden:                 data["IsHidden"],
		IsLive:                   data["IsLive"],
		IsPending:                data["IsPending"],
		IsSold:                   data["IsSold"],
		LocationLatitud:          data["LocationLatitud"],
		LocationLongitude:        data["LocationLongitude"],
		LocationGeocodeCityID:    data["LocationGeocodeCityID"],
		LocationGeocodeCityName1: data["LocationGeocodeCityName1"],
		LocationGeocodeCityName2: data["LocationGeocodeCityName2"],
		LocationGeocodeStateCode: data["LocationGeocodeStateCode"],
	}
}

func ProductIDTolink(productId string) string {
	return fmt.Sprintf("https://www.facebook.com/marketplace/item/%s/", productId)
}

func IsErrorRateLimit(data any) bool {
	errors, exists := GetKey(data, "errors")
	if !exists {
		return false
	}

	errorsList, ok := errors.([]any)
	if !ok {
		return false
	}

	for _, err := range errorsList {
		message, _ := GetKey(err, "message")

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
	id, exists := GetKey(data, "data.viewer.marketplace_product_details_page.target.id")
	if !exists {
		return false
	}
	return id != nil
}

func ProductDetailsIsDeleted(data any) bool {
	details, exists := GetKey(data, "data.viewer.marketplace_product_details_page")
	if exists && details == nil {
		return true
	}
	return false
}

func GetTaxonomiPathJoined(detail any) any {
	taxonomyPath, exists := GetKey(detail, "marketplace_listing_renderable_target.seo_virtual_category.taxonomy_path")
	if !exists {
		return nil
	}

	taxonomiNames := []string{}
	taxonomyPathList, ok := taxonomyPath.([]any)
	if ok {
		for _, taxonomi := range taxonomyPathList {
			taxonomiName, _ := GetKey(taxonomi, "seo_info.seo_url")
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

func GetLocationAttrs(location any) Location {
	locationLatitud, _ := GetKey(location, "latitude")
	locationLongitude, _ := GetKey(location, "longitude")
	locationGeocodeCityName1, _ := GetKey(location, "reverse_geocode.city")
	locationGeocodeCityName2, _ := GetKey(location, "reverse_geocode.city_page.display_name")
	locationGeocodeCityID, _ := GetKey(location, "reverse_geocode.city_page.id")
	locationGeocodeStateCode, _ := GetKey(location, "reverse_geocode.state")
	return Location{
		LocationLatitud:          locationLatitud,
		LocationLongitude:        locationLongitude,
		LocationGeocodeCityName1: locationGeocodeCityName1,
		LocationGeocodeCityName2: locationGeocodeCityName2,
		LocationGeocodeCityID:    locationGeocodeCityID,
		LocationGeocodeStateCode: locationGeocodeStateCode,
	}
}

func ProductDetailsGet(data any) ([]MarketplaceItemDetails, error) {
	detail, exists := GetKey(data, "data.viewer.marketplace_product_details_page")
	if !exists {
		return nil, fmt.Errorf("detail not found")
	}

	productId, exists := GetKey(detail, "target.id")
	if !exists || productId == nil {
		return nil, fmt.Errorf("detail does not have id")
	}

	var products []MarketplaceItemDetails

	productIdLong, _ := GetKey(detail, "target.product_item.id")

	productUrl, _ := GetKey(detail, "target.story.url")
	if productUrl == nil {
		productUrl = ProductIDTolink(productId.(string))
	}
	productTitle, _ := GetKey(detail, "target.marketplace_listing_title")
	productDescription, _ := GetKey(detail, "target.redacted_description.text")
	productPriceAmount, _ := GetKey(detail, "target.listing_price.amount")
	productPriceCurrency, _ := GetKey(detail, "target.listing_price.currency")
	productAttributeData, _ := GetKey(detail, "target.attribute_data")
	productCreation, _ := GetKey(detail, "target.creation_time")

	locationAny, _ := GetKey(detail, "marketplace_listing_renderable_target.location")
	location := GetLocationAttrs(locationAny)

	taxonomiPathJoined := GetTaxonomiPathJoined(detail)

	detailSellerId, _ := GetKey(detail, "target.marketplace_listing_seller.id")
	detailSellerName, _ := GetKey(detail, "target.marketplace_listing_seller.name")
	detailPhotosAny, _ := GetKey(detail, "target.listing_photos")
	detailPhotos := GetPhotosURI(detailPhotosAny)

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
		SellerID:                 detailSellerId,
		SellerName:               detailSellerName,
		Photos:                   detailPhotos,
		LocationLatitud:          location.LocationLatitud,
		LocationLongitude:        location.LocationLongitude,
		LocationGeocodeCityName1: location.LocationGeocodeCityName1,
		LocationGeocodeCityName2: location.LocationGeocodeCityName2,
		LocationGeocodeStateCode: location.LocationGeocodeStateCode,
		LocationGeocodeCityID:    location.LocationGeocodeCityID,
	}

	products = append(products, marketplaceItemDetails)

	return products, nil
}

func ProductsFromSearchValid(data any) bool {
	edges, exists := GetKey(data, "data.marketplace_search.feed_units.edges")
	if !exists {
		return false
	}

	edgesList, ok := edges.([]any)
	if !ok {
		return false
	}

	for _, edge := range edgesList {
		listing, _ := GetKey(edge, "node.listing")
		if listing == nil {
			continue
		}

		productId, exists := GetKey(listing, "id")
		if exists && productId != nil {
			return true
		}
	}

	return false
}

func GetPhotoURI(photo any) any {
	uri, _ := GetKey(photo, "image.uri")
	return uri
	// DO NOT UNSCAPE BECAUSE Marshall scapes it back
	// if uri == nil {
	// 	return nil
	// }
	// uriStr, ok := uri.(string)
	// if !ok {
	// 	return nil
	// }
	// unquoted, err := strconv.Unquote(uriStr)
	// if err != nil {
	// 	return uriStr
	// }
	// return unquoted
}

func GetPhotosURI(photos any) any {
	photosList, ok := photos.([]any)
	if !ok {
		return nil
	}
	photosListURIs := []any{}
	for _, photo := range photosList {
		photosListURIs = append(photosListURIs, GetPhotoURI(photo))
	}
	return photosListURIs
}

func ProductsFromSearchGet(data any) ([]MarketplaceItemDetails, error) {
	edges, exists := GetKey(data, "data.marketplace_search.feed_units.edges")
	if !exists || edges == nil {
		return nil, fmt.Errorf("no marketplace search found")
	}

	edgesList, ok := edges.([]any)
	if !ok {
		return nil, fmt.Errorf("edges is not a list")
	}

	var products []MarketplaceItemDetails

	for _, edge := range edgesList {
		listing, _ := GetKey(edge, "node.listing")
		if listing == nil {
			continue
		}

		productId, exists := GetKey(listing, "id")
		if !exists || productId == nil {
			continue
		}

		productUrl := ProductIDTolink(productId.(string))

		title, _ := GetKey(listing, "marketplace_listing_title")
		price, _ := GetKey(listing, "listing_price.amount")
		time, _ := GetKey(listing, "if_gk_just_listed_tag_on_search_feed.creation_time")

		isHidden, _ := GetKey(listing, "is_hidden")
		isLive, _ := GetKey(listing, "is_live")
		isPending, _ := GetKey(listing, "is_pending")
		isSold, _ := GetKey(listing, "is_sold")

		locationAny, _ := GetKey(listing, "location")
		location := GetLocationAttrs(locationAny)

		sellerId, _ := GetKey(listing, "marketplace_listing_seller.id")
		sellerName, _ := GetKey(listing, "marketplace_listing_seller.name")

		photoPrimaryAny, _ := GetKey(listing, "primary_listing_photo")
		photoPrimary := GetPhotoURI(photoPrimaryAny)

		productDeliveryTypes, _ := GetKey(listing, "delivery_types")

		marketplaceItemDetails := MarketplaceItemDetails{
			ID:                       productId,
			URL:                      productUrl,
			Title:                    title,
			CreationTime:             time,
			PriceAmount:              price,
			LocationLatitud:          location.LocationLatitud,
			LocationLongitude:        location.LocationLongitude,
			LocationGeocodeCityName1: location.LocationGeocodeCityName1,
			LocationGeocodeCityName2: location.LocationGeocodeCityName2,
			LocationGeocodeStateCode: location.LocationGeocodeStateCode,
			LocationGeocodeCityID:    location.LocationGeocodeCityID,
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
	id, exists := GetKey(data, "data.node.entity_id")
	if !exists {
		return false
	}
	return id != nil
}

func ProducFromDataGet(data any) ([]MarketplaceItemDetails, error) {
	node, exists := GetKey(data, "data.node")
	if !exists {
		return nil, fmt.Errorf("data node not found")
	}

	productId, exists := GetKey(node, "entity_id")
	if !exists || productId == nil {
		return nil, fmt.Errorf("product does not have id")
	}

	var products []MarketplaceItemDetails

	productIdLong, _ := GetKey(node, "data.product_item_id")

	productTitle, _ := GetKey(node, "data.title")

	productCategory, _ := GetKey(node, "data.upsell_type")

	// do not save this caterogy
	if productCategory != nil && productCategory == "CATEGORY_MISCELLANEOUS_UPSELL" {
		productCategory = nil
	}

	productUrl := ProductIDTolink(productId.(string))

	productPriceCurrency, _ := GetKey(node, "data.price.currency")

	locationAny, _ := GetKey(node, "entity.location")
	location := GetLocationAttrs(locationAny)

	productCreation, _ := GetKey(node, "listing.creation_time")

	marketplaceItemDetails := MarketplaceItemDetails{
		ID:                       productId,
		URL:                      productUrl,
		IDLong:                   productIdLong,
		Category:                 productCategory,
		Title:                    productTitle,
		PriceCurrency:            productPriceCurrency,
		LocationLatitud:          location.LocationLatitud,
		LocationLongitude:        location.LocationLongitude,
		LocationGeocodeCityName1: location.LocationGeocodeCityName1,
		LocationGeocodeCityName2: location.LocationGeocodeCityName2,
		LocationGeocodeStateCode: location.LocationGeocodeStateCode,
		LocationGeocodeCityID:    location.LocationGeocodeCityID,
		CreationTime:             productCreation,
	}

	products = append(products, marketplaceItemDetails)

	return products, nil
}
