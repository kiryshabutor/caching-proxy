package cache

import (
	"sync"
)

type CacheResponse struct {
	Body []byte
	Status int
	Header map[string][]string
}

type Cache struct {
	data map[string]CacheResponse	
	mu sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]CacheResponse),
	}
}

func (c *Cache) Get(key string) (CacheResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	response, ok := c.data[key]
	return response, ok
}

func (c *Cache) Set(key string, response CacheResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data[key] = response
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data = make(map[string]CacheResponse)
}


