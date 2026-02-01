package main

import (
	"CachingProxy/pkg/cache"
	"CachingProxy/pkg/server"
	"flag"
	"fmt"
	"log"
)

func main() {
	port := flag.Int("port", 2020, "port to listen on")
	cacheSize := flag.Int("cache-size", 100, "max number of items in cache")
	origin := flag.String("origin", "http://localhost:7070", "url to proxy to")
	clearCache := flag.Bool("clear-cache", false, "clear cache")
	flag.Parse()

	cache := cache.NewCache(*cacheSize)
	server := server.NewServer(*port, *origin, cache)

	if *clearCache {
		cache.Clear()
		log.Println("Cache cleared")
		return
	}

	fmt.Println("Starting server on port", *port)
	if err := server.Run(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}