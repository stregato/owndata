package core

import (
	"sync"
)

// BMap is a generic bidirectional map.
type BMap[K comparable, V comparable] struct {
	lock       sync.RWMutex
	keyToValue map[K]V
	valueToKey map[V]K
}

// NewBMap creates and returns a new instance of a BMap.
func NewBMap[K comparable, V comparable]() *BMap[K, V] {
	return &BMap[K, V]{
		keyToValue: make(map[K]V),
		valueToKey: make(map[V]K),
	}
}

// Set adds or updates the key-value pair in the bimap.
func (bm *BMap[K, V]) Set(key K, value V) {
	bm.lock.Lock()
	defer bm.lock.Unlock()
	// Ensure uniqueness
	if existingValue, ok := bm.keyToValue[key]; ok {
		delete(bm.valueToKey, existingValue)
	}
	if existingKey, ok := bm.valueToKey[value]; ok {
		delete(bm.keyToValue, existingKey)
	}
	bm.keyToValue[key] = value
	bm.valueToKey[value] = key
}

// FromMap takes a normal map and converts it to a BMap.
func FromMap[K comparable, V comparable](m map[K]V) *BMap[K, V] {
	bMap := NewBMap[K, V]()
	for key, value := range m {
		bMap.Set(key, value)
	}
	return bMap
}
