package cache

import (
	"fmt"
	"time"
)

/*
*********************
    Private Funtions
*********************
*/

func (cache *Cache) gcLoop() {
    ticker := time.NewTicker(cache.gcInterval)
    for {
        select {
        case <-ticker.C:
            fmt.Println("----------- GC START ----------: ", len(cache.items))
            cache.Flush()
        case <-cache.stopGc:
            ticker.Stop()
            return
        }
    }
}

func (cache *Cache) delete(key string) {
    delete(cache.items, key)
}

/*
*********************
    Public Funtions
*********************
*/

// StopGc is using for stopping the GC
func (cache *Cache) StopGc() {
    cache.stopGc <- true
}

// StartGc is using for starting the GC
func (cache *Cache) StartGc() {
    go cache.gcLoop()
}

// Expired is using for identify the expired records
func (item Item) Expired() bool {
    if item.Expiration == 0 {
        return false
    }
    return time.Now().UnixNano() > item.Expiration
}

// Delete is using for delete a cache record
func (cache *Cache) Delete(k string) {
    cache.mu.Lock()
    cache.delete(k)
    defer cache.mu.Unlock()
}

// DeleteExpired is using for deleting the expired records
func (cache *Cache) DeleteExpired() {
    now := time.Now().UnixNano()
    cache.mu.Lock()
    for key, value := range cache.items {
        if value.Expiration > 0 && now > value.Expiration {
            cache.delete(key)
        }
    }
    defer cache.mu.Unlock()
}

// NewCache is using for creating a new cache entity
func NewCache(defaultExpiration, gcInterval time.Duration) *Cache {
    cache := &Cache{
        items:                  map[string]Item{},
        defaultExpiration:      defaultExpiration,
        gcInterval:             gcInterval,
        stopGc:                 make(chan bool),
    }

    go cache.gcLoop()
    return cache
}

// Set is for adding the record to the cache
func (cache *Cache) Set(key string, value interface{}, duration time.Duration) {
    var expiration int64
    if duration == DefaultExpiration {
        duration = cache.defaultExpiration
    }
    if duration > 0 {
        expiration = time.Now().Add(duration).UnixNano()
    }
    cache.mu.Lock()
    cache.items[key] = Item{
        Object:     value,
        Expiration: expiration,
    }
    defer cache.mu.Unlock()
}

// Get is for getting data
func (cache *Cache) Get(key string) (interface{}) {

    cache.mu.RLock()
    item, found := cache.items[key]
    if !found {
        defer cache.mu.RUnlock()
        return nil
    }
    if item.Expired() {
        return nil
    }
    defer cache.mu.RUnlock()
    return item.Object
}

// Flush is empty cache
func (cache *Cache) Flush() {
    cache.mu.Lock()
    defer cache.mu.Unlock()
    cache.items = map[string]Item{}
}
