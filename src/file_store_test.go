package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStoreGet(t *testing.T) {
	tmpDir := t.TempDir()

	productID := "test-product-123"
	store := NewProductFileStore(productID)
	store.SetDir(tmpDir)

	// Create test data
	testData := MarketplaceItemDetails{
		ID:            "123",
		Title:         "Test Product",
		Description:   "Test Description",
		PriceAmount:   99.99,
		PriceCurrency: "USD",
		SellerID:      "seller-456",
		SellerName:    "Test Seller",
	}

	err := store.Save(testData)
	require.NoError(t, err)

	store2 := NewProductFileStore(productID)
	store2.SetDir(tmpDir)

	gotData, err := store2.Get()
	require.NoError(t, err)

	assert.NotNil(t, gotData)

	assert.Equal(t, testData.ID, gotData.ID)
	assert.Equal(t, testData.Title, gotData.Title)
	assert.Equal(t, testData.Description, gotData.Description)
	assert.Equal(t, testData.PriceAmount, gotData.PriceAmount)
	assert.Equal(t, testData.PriceCurrency, gotData.PriceCurrency)
	assert.Equal(t, testData.SellerID, gotData.SellerID)
	assert.Equal(t, testData.SellerName, gotData.SellerName)
}

func TestFileStoreSave(t *testing.T) {
	tmpDir := t.TempDir()

	productID := "test-product-456"

	store := NewProductFileStore(productID)
	store.SetDir(tmpDir)

	data, err := store.Get()

	assert.NoError(t, err)
	assert.NotNil(t, data)

	assert.Equal(t, nil, data.ID)
	assert.Equal(t, nil, data.Title)
	assert.Equal(t, nil, data.Description)

	testData := MarketplaceItemDetails{
		ID:          "123",
		Title:       "Test Product",
		Description: "Test Description",
	}

	store.Save(testData)

	store2 := NewProductFileStore(productID)
	store2.SetDir(tmpDir)

	data2, err := store2.Get()

	assert.NoError(t, err)
	assert.NotNil(t, data2)

	assert.Equal(t, testData.ID, data2.ID)
	assert.Equal(t, testData.Title, data2.Title)
	assert.Equal(t, testData.Description, data2.Description)

	// update
	testData.ID = "321"

	store2.Save(testData)

	data3, err := store2.Get()
	assert.Equal(t, "321", data3.ID)
}
