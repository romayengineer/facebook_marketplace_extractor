package main

import (
	"fmt"
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

func GetProductDetails(data any) (*MarketplaceItemDetails, error) {
	detail := GetKey(data, "data.viewer.marketplace_product_details_page")
	if detail == nil {
		return nil, fmt.Errorf("detail not found")
	}

	productId := GetKey(detail, "target.id")
	if productId == nil {
		return nil, fmt.Errorf("detail does not have id")
	}
	productIdLong := GetKey(detail, "target.product_item.id")

	productUrl := GetKey(detail, "target.story.url")
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

	detailSellerId := GetKey(detail, "target.marketplace_listing_seller.id")
	detailSellerName := GetKey(detail, "target.marketplace_listing_seller.name")
	detailPhotos := GetKey(detail, "target.listing_photos")

	marketplaceItemDetails := MarketplaceItemDetails{
		ID:                       productId,
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

	store := NewProductFileStore(productId.(string))
	newData, _ := store.Save(marketplaceItemDetails)

	return newData, nil
}

func GetProductsFromSearch(data any) ([]*MarketplaceItemDetails, error) {
	edges := GetKey(data, "data.marketplace_search.feed_units.edges")
	if edges == nil {
		return nil, fmt.Errorf("no marketplace search found")
	}

	edgesList, ok := edges.([]any)
	if !ok {
		return nil, fmt.Errorf("edges is not a list")
	}

	var products []*MarketplaceItemDetails
	for _, edge := range edgesList {
		listing := GetKey(edge, "node.listing")
		if listing == nil {
			continue
		}

		productId := GetKey(listing, "id")
		if productId == nil {
			continue
		}

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

		store := NewProductFileStore(productId.(string))
		product, _ := store.Save(marketplaceItemDetails)

		products = append(products, product)
	}

	return products, nil
}

func GetProducFromData(data any) (*MarketplaceItemDetails, error) {
	node := GetKey(data, "data.node")
	if node == nil {
		return nil, fmt.Errorf("data node not found")
	}

	productId := GetKey(node, "entity_id")
	if productId == nil {
		return nil, fmt.Errorf("product does not have id")
	}

	productIdLong := GetKey(node, "data.product_item_id")

	productTitle := GetKey(node, "data.title")

	productCategory := GetKey(node, "data.upsell_type")

	productPriceCurrency := GetKey(node, "data.price.currency")

	location := GetKey(node, "entity.location")
	cityName1 := GetKey(location, "reverse_geocode.city")
	cityName2 := GetKey(location, "reverse_geocode.city_page.display_name")
	cityID := GetKey(location, "reverse_geocode.city_page.id")
	stateCode := GetKey(location, "reverse_geocode.state")

	productCreation := GetKey(node, "listing.creation_time")

	marketplaceItemDetails := MarketplaceItemDetails{
		ID:                       productId,
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

	store := NewProductFileStore(productId.(string))
	newData, _ := store.Save(marketplaceItemDetails)

	return newData, nil
}
