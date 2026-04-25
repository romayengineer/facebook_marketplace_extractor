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
	ID            any
	URL           any
	Title         any
	Description   any
	PriceAmount   any
	PriceCurrency any
	AttributeData any
	CreationTime  any
	Location      any
	SellerID      any
	SellerName    any
	Photos        any
}

func NewMarketplaceItemDetails(
	id any,
	url any,
	title any,
	description any,
	priceAmount any,
	priceCurrency any,
	attributeData any,
	creationTime any,
	location any,
	sellerId any,
	sellerName any,
	photos any,
) MarketplaceItemDetails {
	return MarketplaceItemDetails{
		ID:            id,
		URL:           url,
		Description:   description,
		AttributeData: attributeData,
		Title:         title,
		CreationTime:  creationTime,
		Location:      location,
		PriceAmount:   priceAmount,
		PriceCurrency: priceCurrency,
		SellerID:      sellerId,
		SellerName:    sellerName,
		Photos:        photos,
	}
}

func GetProductDetails(data any) (*MarketplaceItemDetails, error) {
	detail := GetKey(data, "data.viewer.marketplace_product_details_page")
	if detail == nil {
		return nil, fmt.Errorf("detail not found")
	}

	detailId := GetKey(detail, "target.id")
	if detailId == nil {
		return nil, fmt.Errorf("detail does not have id")
	}

	detailDescription := GetKey(detail, "target.redacted_description.text")
	detailAttributeData := GetKey(detail, "target.attribute_data")
	detailTitle := GetKey(detail, "target.marketplace_listing_title")
	detailCreation := GetKey(detail, "target.creation_time")
	// detailLocation := GetKey(detail, "target.item_location")
	detailLocation := GetKey(detail, "marketplace_listing_renderable_target.location")
	detailPriceAmount := GetKey(detail, "target.listing_price.amount")
	detailPriceCurrency := GetKey(detail, "target.listing_price.currency")
	detailSellerId := GetKey(detail, "target.marketplace_listing_seller.id")
	detailSellerName := GetKey(detail, "target.marketplace_listing_seller.name")
	detailPhotos := GetKey(detail, "target.listing_photos")
	detailUrl := GetKey(detail, "story.shareable.url")

	marketplaceItemDetails := NewMarketplaceItemDetails(
		detailId,
		detailUrl,
		detailTitle,
		detailDescription,
		detailPriceAmount,
		detailPriceCurrency,
		detailAttributeData,
		detailCreation,
		detailLocation,
		detailSellerId,
		detailSellerName,
		detailPhotos,
	)

	store := NewProductFileStore(detailId.(string))
	newData, _ := store.Save(marketplaceItemDetails)

	return newData, nil
}
