package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

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
