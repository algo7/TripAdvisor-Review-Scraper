package database

import (
	"container_provisioner/utils"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
)

// SetCache sets the given value to the r2 storage object list key
func SetCache(key string, value []byte) {
	ctx := context.Background()
	// Timeout set to 5 minutes
	err := rdb.Set(ctx, key, value, time.Minute*5).Err()
	utils.ErrorHandler(err)
}

// CacheLookUp checks if the given value exists in the cache, returns the value if it exists
func CacheLookUp(key string) string {

	ctx := context.Background()

	cachedObjectsList, err := rdb.Get(ctx, key).Result()

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
