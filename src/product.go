package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

type FileStoreImpl[T any] struct {
	id       string
	filePath string
	data     *T
}

type FileStore[T any] interface {
	Get() error
	Save(data T) error
}

func NewProductFileStore(productId string) FileStore[MarketplaceItemDetails] {
	fileName := fmt.Sprintf("detail_%v.json", productId)
	filePath := filepath.Join("data", fileName)
	return &FileStoreImpl[MarketplaceItemDetails]{id: productId, filePath: filePath}
}

func (pfs *FileStoreImpl[T]) Get() error {
	content, err := os.ReadFile(pfs.filePath)
	if err != nil {
		return err
	}

	var data T
	if err := json.Unmarshal(content, &data); err == nil {
		return err
	}

	pfs.data = &data

	return nil
}

func (pfs *FileStoreImpl[T]) Save(data T) error {
	if pfs.data == nil {
		err := pfs.Get()
		if err != nil {
			return err
		}
	}

	pfsValue := reflect.ValueOf(pfs.data).Elem()
	dataValue := reflect.ValueOf(data)

	for i := 0; i < pfsValue.NumField(); i++ {
		pfsField := pfsValue.Field(i)
		dataField := dataValue.Field(i)

		if !dataField.IsNil() {
			pfsField.Set(dataField)
		}
	}

	indented, err := json.MarshalIndent(pfs.data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(pfs.filePath, indented, 0644); err != nil {
		return err
	}

	return nil
}

func GetProductDetails(data any) (*MarketplaceItemDetails, error) {
	detail, err := GetKey(data, "data.viewer.marketplace_product_details_page")
	if err != nil {
		return nil, err
	}
	detailId, err := GetKey(detail, "target.id")
	if err != nil {
		return nil, err
	}
	detailDescription, err := GetKey(detail, "target.redacted_description.text")
	if err != nil {
		return nil, err
	}
	detailAttributeData, err := GetKey(detail, "target.attribute_data")
	if err != nil {
		return nil, err
	}
	detailTitle, err := GetKey(detail, "target.marketplace_listing_title")
	if err != nil {
		return nil, err
	}
	detailCreation, err := GetKey(detail, "target.creation_time")
	if err != nil {
		return nil, err
	}
	// detailLocation, err := GetKey(detail, "target.item_location")
	detailLocation, err := GetKey(detail, "marketplace_listing_renderable_target.location")
	if err != nil {
		return nil, err
	}
	detailPriceAmount, err := GetKey(detail, "target.listing_price.amount")
	if err != nil {
		return nil, err
	}
	detailPriceCurrency, err := GetKey(detail, "target.listing_price.currency")
	if err != nil {
		return nil, err
	}
	detailSellerId, err := GetKey(detail, "target.marketplace_listing_seller.id")
	if err != nil {
		return nil, err
	}
	detailSellerName, err := GetKey(detail, "target.marketplace_listing_seller.name")
	if err != nil {
		return nil, err
	}

	// optional
	detailPhotos, err := GetKey(detail, "target.listing_photos")

	marketplaceItemDetails := NewMarketplaceItemDetails(
		detailId,
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

	filename := fmt.Sprintf("detail_%v.json", detailId)

	indented, err := json.MarshalIndent(marketplaceItemDetails, "", "  ")

	if err := os.WriteFile(filepath.Join("data", filename), indented, 0644); err != nil {
		return nil, err
	}

	return &marketplaceItemDetails, nil
}
