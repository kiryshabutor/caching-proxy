package cache

import (
	"container/list"
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

type Entry struct {
	Key string
	Value CacheResponse
}

type Cache struct {
	data map[string]*list.Element
	cacheSize int
	cacheCounter int
	list *list.List
	mu   sync.RWMutex
}

func NewCache(cacheSize int) *Cache {
	c := &Cache{
		data: make(map[string]*list.Element),
		cacheSize: cacheSize,
		list: list.New(),
	}
	c.loadFromFile()
	return c
}

func (c *Cache) Get(key string) (CacheResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.data[key]
	if ok {
		c.list.MoveToFront(elem)
		return elem.Value.(Entry).Value, true
	}
	return CacheResponse{}, false
}

func (c *Cache) Set(key string, response CacheResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[key]; ok {
		return
	}
    
	for c.cacheCounter >= c.cacheSize {
		c.deleteLeastUsed()
	}

	c.data[key] = c.list.PushFront(Entry{Key: key, Value: response})
	c.cacheCounter++
	c.saveToFile()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*list.Element)
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Error clearing cache file: %v", err)
	}
}

func (c *Cache) deleteLeastUsed() {
	lastEl := c.list.Back()
	if lastEl != nil {
		delete(c.data, lastEl.Value.(Entry).Key)
		c.list.Remove(lastEl)
	}
	c.cacheCounter--
}

func (c *Cache) saveToFile() {
	file, err := os.Create(cacheFile)
	if err != nil {
		log.Printf("Error creating cache file: %v", err)
		return
	}
	defer file.Close()
	entries := make([]Entry, 0, len(c.data))
	for e := c.list.Front(); e != nil; e = e.Next() {
		entries = append(entries, e.Value.(Entry))
	}
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(entries); err != nil {
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

	var entries []Entry	

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		log.Printf("Error decoding cache data: %v", err)
	}
	for _, e := range entries {
		c.data[e.Key] = c.list.PushBack(e)
		c.cacheCounter++
	}
}
