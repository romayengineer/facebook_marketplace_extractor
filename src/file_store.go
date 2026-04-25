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
	fileName string
	fileDir  string
	data     *T
}

type FileStore[T any] interface {
	SetDir(dir string)
	Get() (*T, error)
	Save(data T) (*T, error)
}

func NewProductFileStore(productId string) FileStore[MarketplaceItemDetails] {
	fileName := fmt.Sprintf("detail_%v.json", productId)
	return &FileStoreImpl[MarketplaceItemDetails]{id: productId, fileName: fileName, fileDir: "data"}
}

func (pfs *FileStoreImpl[T]) SetDir(dir string) {
	pfs.fileDir = dir
}

func (pfs *FileStoreImpl[T]) Get() (*T, error) {
	var data T

	filePath := filepath.Join(pfs.fileDir, pfs.fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		pfs.data = &data
		return pfs.data, nil
	}

	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	pfs.data = &data

	return pfs.data, nil
}

func (pfs *FileStoreImpl[T]) Save(data T) (*T, error) {
	if pfs.data == nil {
		_, err := pfs.Get()
		if err != nil {
			return nil, err
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
		return nil, err
	}

	filePath := filepath.Join(pfs.fileDir, pfs.fileName)
	if err := os.WriteFile(filePath, indented, 0644); err != nil {
		return nil, err
	}

	return pfs.data, nil
}
