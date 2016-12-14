// Package lru lru with TTL
package lru

import (
	"errors"
	"sync"
	"time"

	hlru "github.com/hashicorp/golang-lru"
)

// LruWithTTL lru with ttl
type LruWithTTL struct {
	hlru.Cache
	schedule      map[interface{}]bool
	scheduleMutex sync.Mutex
}

// NewTTL creates an LRU of the given size
func NewTTL(size int) (*LruWithTTL, error) {
	return NewTTLWithEvict(size, nil)
}

// NewTTLWithEvict creates an LRU of the given size and a evict callback function
func NewTTLWithEvict(size int, onEvicted func(key interface{}, value interface{})) (*LruWithTTL, error) {
	if size <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	c, err := hlru.NewWithEvict(size, onEvicted)
	if err != nil {
		return nil, err
	}
	return &LruWithTTL{*c, make(map[interface{}]bool), sync.Mutex{}}, nil
}

func (lru *LruWithTTL) clearSchedule(key interface{}) {
	lru.scheduleMutex.Lock()
	defer lru.scheduleMutex.Unlock()
	delete(lru.schedule, key)
}

// AddWithTTL add an key:val with TTL
func (lru *LruWithTTL) AddWithTTL(key, value interface{}, ttl time.Duration) bool {
	lru.scheduleMutex.Lock()
	defer lru.scheduleMutex.Unlock()
	if lru.schedule[key] {
		// already scheduled, nothing to do
	} else {
		lru.schedule[key] = true
		// Schedule cleanup
		go func() {
			defer lru.Cache.Remove(key)
			defer lru.clearSchedule(key)
			time.Sleep(ttl)
		}()
	}
	return lru.Cache.Add(key, value)
}
