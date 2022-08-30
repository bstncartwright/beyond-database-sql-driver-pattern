package main

import (
	"fmt"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/02/cache"
)

func main() {
	// create cache to get data from

	// change this to change the cache type
	cacheType := "redis"

	var repo cache.Cache

	switch cacheType {
	case "redis":
		repo = cache.NewRedis()
	default:
		repo = cache.NewInMemory()
	}

	data, ok := repo.Get("data")
	if !ok {
		panic("not found in cache")
	}

	fmt.Printf("data: %v\n", data)
}
