package main

import (
	"fmt"
	"log"
)

type OrderedMap struct {
	data  map[string]string
	order []string
}

func (om *OrderedMap) Set(key string, value string) {
	if _, exists := om.data[key]; !exists {
		om.order = append(om.order, key)
	}
	om.data[key] = value
}

func (om *OrderedMap) Get(key string) (string, bool) {
	data, exists := om.data[key]
	return data, exists
}

func (om *OrderedMap) GetDefault(key string, def string) string {
	val, exists := om.Get(key)
	if exists {
		return val
	}
	return def
}

func (om *OrderedMap) Keys() []string {
	return om.order
}

func (om *OrderedMap) Print() {
	for _, key := range om.Keys() {
		value, _ := om.Get(key)
		log.Printf("OrderedMap.Print: %s = %s", key, value)
	}
}

func (om *OrderedMap) Compare(om2 OrderedMap) {
	var v1 string
	var v2 string
	var exists bool
	equal := 0
	for _, k := range om.Keys() {
		if v2, exists = om2.data[k]; !exists {
			fmt.Printf("key changed %s\n", k)
		}
		if v1, exists = om.data[k]; !exists {
			fmt.Printf("key changed %s\n", k)
		}
		if v1 != v2 {
			fmt.Printf("key changed %s\n", k)
		} else {
			equal += 1
		}
	}
	fmt.Printf("keys equal: %02d\n\n\n", equal)
}

func NewOrderedMap() OrderedMap {
	return OrderedMap{
		data:  map[string]string{},
		order: []string{},
	}
}
