package database

import (
	"container_provisioner/utils"
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
)

// CacheLookUp checks if the r2 storage object list exists in the cache
func CacheLookUp() string {

	ctx := context.Background()

	cachedObjectsList, err := rdb.Get(ctx, "r2StorageObjectsList").Result()

	// If the key does not exist, return an empty string
	if err == redis.Nil {
		return ""
	}

	// If actual error
	if err != nil {
		utils.ErrorHandler(err)
	}

	// If the key exists, return the value
	return cachedObjectsList
}
