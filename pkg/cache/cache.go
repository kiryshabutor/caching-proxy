package cache

import (
	"encoding/gob"
	"log"
	"os"
	"sync"
)

const cacheFile = "cache.gob"

type CacheResponse struct {
	Body   []byte
	Status int
	Header map[string][]string
}

type Cache struct {
	data map[string]CacheResponse
	mu   sync.RWMutex
}

func NewCache() *Cache {
	c := &Cache{
		data: make(map[string]CacheResponse),
	}
	c.loadFromFile()
	return c
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
	c.saveToFile()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]CacheResponse)
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Error clearing cache file: %v", err)
	}
}

func (c *Cache) saveToFile() {
	file, err := os.Create(cacheFile)
	if err != nil {
		log.Printf("Error creating cache file: %v", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(c.data); err != nil {
		log.Printf("Error encoding cache data: %v", err)
	}
}

func (c *Cache) loadFromFile() {
	file, err := os.Open(cacheFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error opening cache file: %v", err)
		}
		return
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&c.data); err != nil {
		log.Printf("Error decoding cache data: %v", err)
	}
}
