package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"

	"github.com/redis/go-redis/v9"
)

// RedisClient struct to hold the redis client
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new redis client. The redis host address is fetched from the environment variable REDIS_HOST. If the environment variable is not set, it defaults to "localhost:6379". The redis password is fetched from the environment variable REDIS_PASS. If the environment variable is not set, no password is used.
func NewRedisClient() *RedisClient {
	// Get the redis host address from the environment variable
	redisHost := os.Getenv("REDIS_HOST")

	// If the environment variable is not set, use the default address
	if redisHost == "" {
		redisHost = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,               // use the given Addr or default Addr
		Password: os.Getenv("REDIS_PASS"), // no password set
		DB:       0,                       // use default DB
	})

	return &RedisClient{
		Client: rdb,
	}
}

// SetCache store the given value in redis
func (r *RedisClient) SetCache(key string, value any) {

	// Encode the slice of Row structs into a byte slice
	encodedValue, err := json.Marshal(value)
	utils.ErrorHandler(err)

	ctx := context.Background()
	// Timeout set to 5 minutes
	err = r.Client.Set(ctx, key, string(encodedValue), time.Minute*1).Err()
	utils.ErrorHandler(err)
}

// CacheLookUp checks if the given value exists in the cache, returns the value if it exists
func (r *RedisClient) CacheLookUp(key string) string {

	ctx := context.Background()

	cachedObjectsList, err := r.Client.Get(ctx, key).Result()

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
func (r *RedisClient) CheckConnection() (string, error) {

	ctx := context.Background()

	// Ping the redis server to check if it is up
	resp, err := r.Client.Ping(ctx).Result()
	if err != nil {
		return "", fmt.Errorf("failed to ping redis server: %w", err)
	}

	return resp, nil

}

// SetLock sets a lock on the given key
func (r *RedisClient) SetLock(key string) bool {

	ctx := context.Background()

	lockSuccess, err := r.Client.SetNX(ctx, key, "1", 0).Result()

	if err != nil {
		return false
	}

	return lockSuccess
}

// ReleaseLock releases the lock on the given key
func (r *RedisClient) ReleaseLock(key string) error {

	ctx := context.Background()

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}
