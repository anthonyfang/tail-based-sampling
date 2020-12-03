package cache

import (
    "sync"
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
            cache.DeleteExpired()
        case <-cache.stopGc:
            ticker.Stop()
            return
        }
    }
}

func (cache *Cache) delete(key string) {
    cache.items.Delete(key)
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
    cache.delete(k)
}

// DeleteExpired is using for deleting the expired records
func (cache *Cache) DeleteExpired() {
    now := time.Now().UnixNano()
    cache.items.Range(func(key interface{}, value interface{}) bool {
        value1 := value.(Item)
        if value1.Expiration > 0 && now > value1.Expiration {
            cache.items.Delete(key)
        }
        return true
    })
}

// NewCache is using for creating a new cache entity
func NewCache(defaultExpiration, gcInterval time.Duration) *Cache {
    cache := &Cache{
        items:                  sync.Map{},
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
    cache.items.Store(key, Item{
        Object:     value,
        Expiration: expiration,
    })
}

// Get is for getting data
func (cache *Cache) Get(key string) (interface{}) {
    value, found := cache.items.Load(key)
    if !found {
        return nil
    }
    item := value.(Item)
    if item.Expired() {
        return nil
    }
    return item.Object
}

// // Flush is empty cache
// func (cache *Cache) Flush() {
//     cache.mu.Lock()
//     defer cache.mu.Unlock()
//     cache.items = map[string]Item{}
// }
