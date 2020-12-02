package cache

import (
    "sync"
    "time"
)

// Item is the cache record struct
type Item struct {
    Object     interface{}
    Expiration int64
}

const (
    // DefaultExpiration is using for setting the expire time
    DefaultExpiration   time.Duration = 0
)

// Cache is the cache entity structure 
type Cache struct {
    items               map[string]Item
    mu                  sync.RWMutex
    defaultExpiration   time.Duration
    gcInterval          time.Duration
    stopGc              chan bool
}
