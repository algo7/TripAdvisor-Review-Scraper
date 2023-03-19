package database

import (
	"container_provisioner/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb = redis.NewClient(&redis.Options{
		Addr:     getRedisHostAddress(), // use the given Addr or default Addr
		Password: "",                    // no password set
		DB:       0,                     // use default DB
	})
)

// SetCache store the given value in redis
func SetCache(key string, value any) {

	// Encode the slice of Row structs into a byte slice
	encodedValue, err := json.Marshal(value)
	utils.ErrorHandler(err)

	ctx := context.Background()
	// Timeout set to 5 minutes
	err = rdb.Set(ctx, key, string(encodedValue), time.Minute*5).Err()
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

// RedisConnectionCheck checks if the redis server is up and running
func RedisConnectionCheck() {

	ctx := context.Background()

	// Ping the redis server to check if it is up
	resp, err := rdb.Ping(ctx).Result()
	utils.ErrorHandler(err)

	fmt.Println("Redis connection established", resp)
}

// getRedisHostAddress checks if custom redis host address is supplied, if not, returns the default address
func getRedisHostAddress() string {

	// Get the redis host address from the environment variable
	redisHost := os.Getenv("REDIS_HOST")

	// If the environment variable is not set, use the default address
	if redisHost == "" {
		redisHost = "localhost:6379"
	}

	return redisHost
}
