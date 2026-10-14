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

// ProductDetailsValid reports whether the GraphQL payload is a Marketplace PDP
// (product detail page) response. See ProductDetailsGet for the requirements to
// obtain a PDP response containing `redacted_description`.
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

func GetLocationAttrs(location any) Location {
	return Location{
		LocationLatitud:          GetKey(location, "latitude"),
		LocationLongitude:        GetKey(location, "longitude"),
		LocationGeocodeCityName1: GetKey(location, "reverse_geocode.city"),
		LocationGeocodeCityName2: GetKey(location, "reverse_geocode.city_page.display_name"),
		LocationGeocodeCityID:    GetKey(location, "reverse_geocode.city_page.id"),
		LocationGeocodeStateCode: GetKey(location, "reverse_geocode.state"),
	}
}

// ProductDetailsGet maps a Facebook Marketplace PDP (product detail page) GraphQL
// response to MarketplaceItemDetails.
//
// The field we care about most is Description, which Facebook returns as
// `redacted_description` (the lead text shown at the top of a listing). Obtaining it
// has hard requirements:
//
//  1. Authenticated session. The request must run against a logged-in Facebook
//     session; otherwise the PDP response is a login wall with no description.
//     In practice this means launching with a Chrome profile that already has the
//     session (FME_USER_DATA_DIR) - see docs/redacted_description.md.
//
//  2. A real PDP request must be triggered. `MarketplacePDPContainerQuery` is only
//     sent when a listing is actually opened in the browser via a client-side click
//     on a search result. A plain URL navigation (Goto) does NOT reliably trigger it,
//     which is why Pages.OpenFirstProduct clicks the first item link after the search.
//
//  3. The response must be intercepted. ContextEventHandlers.OnResponse captures
//     requests to https://www.facebook.com/api/graphql/ and writes the body to
//     data/MarketplacePDPContainerQuery/*.json (the friendly name is read from the
//     request post-data field `fb_api_req_friendly_name`).
//
//  4. The description is read from the dotted path:
//     data.viewer.marketplace_product_details_page.target.redacted_description.text
//
//  5. ProcessData merges non-nil fields into data/__Details/detail_<id>.json, so the
//     description is added to the listing record already created from the search feed.
//
// The same intercepted request also provides the reusable headers + post-data that
// RunRequestDecompress replays (with variables.targetId replaced) to fetch details for
// the other listings.
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

	location := GetLocationAttrs(GetKey(detail, "marketplace_listing_renderable_target.location"))

	taxonomiPathJoined := GetTaxonomiPathJoined(detail)

	detailSellerId := GetKey(detail, "target.marketplace_listing_seller.id")
	detailSellerName := GetKey(detail, "target.marketplace_listing_seller.name")
	detailPhotos := GetPhotosURI(GetKey(detail, "target.listing_photos"))

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

func GetPhotoURI(photo any) any {
	uri := GetKey(photo, "image.uri")
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

		location := GetLocationAttrs(GetKey(listing, "location"))

		sellerId := GetKey(listing, "marketplace_listing_seller.id")
		sellerName := GetKey(listing, "marketplace_listing_seller.name")

		photoPrimary := GetPhotoURI(GetKey(listing, "primary_listing_photo"))

		productDeliveryTypes := GetKey(listing, "delivery_types")

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

	location := GetLocationAttrs(GetKey(node, "entity.location"))

	productCreation := GetKey(node, "listing.creation_time")

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
