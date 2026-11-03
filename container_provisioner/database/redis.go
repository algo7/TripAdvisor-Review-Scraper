package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

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
func (r *RedisClient) SetCache(key string, value any) error {

	// Encode the slice of Row structs into a byte slice
	encodedValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	ctx := context.Background()
	// Timeout set to 5 minutes
	err = r.Client.Set(ctx, key, string(encodedValue), time.Minute*1).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}

	return nil
}

// CacheLookUp checks if the given value exists in the cache, returns the value if it exists
func (r *RedisClient) CacheLookUp(key string) (string, error) {

	ctx := context.Background()

	cachedObjectsList, err := r.Client.Get(ctx, key).Result()

	// If the key does not exist, return an empty string
	if err == redis.Nil {
		return "", nil
	}

	// If actual error
	if err != nil {
		return "", fmt.Errorf("failed to get value from redis: %w", err)
	}

	// If the key exists, return the value
	return cachedObjectsList, nil
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

// AcquireLock takes the lock on key for owner if nobody holds it. The lock expires after ttl unless it is renewed.
func (r *RedisClient) AcquireLock(key string, owner string, ttl time.Duration) (bool, error) {

	ctx := context.Background()

	err := r.Client.SetArgs(ctx, key, owner, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Err()

	// SET NX replies with nil when the key already exists
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock %s: %w", key, err)
	}

	return true, nil
}

// RenewLock resets the expiry of the lock on key to ttl if owner still holds it.
// It returns false when the lock expired or now belongs to someone else.
// Needs Redis 8.4 or newer for SET IFEQ.
func (r *RedisClient) RenewLock(key string, owner string, ttl time.Duration) (bool, error) {

	ctx := context.Background()

	// SET IFEQ only writes when the current value is owner, in one atomic step.
	// It never recreates an expired lock, because a missing key doesn't match.
	err := r.Client.SetIFEQ(ctx, key, owner, owner, ttl).Err()

	// SET IFEQ replies with nil when the value doesn't match
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to renew lock %s: %w", key, err)
	}

	return true, nil
}

// ReleaseLock deletes the lock on key if owner still holds it, so a lock taken over by someone else is left alone.
// Needs Redis 8.4 or newer for DELEX.
func (r *RedisClient) ReleaseLock(key string, owner string) (bool, error) {

	ctx := context.Background()

	// DELEX IFEQ only deletes when the current value is owner, in one atomic step
	deleted, err := r.Client.DelExArgs(ctx, key, redis.DelExArgs{
		Mode:       "IFEQ",
		MatchValue: owner,
	}).Result()
	if err != nil {
		return false, fmt.Errorf("failed to release lock %s: %w", key, err)
	}

	return deleted == 1, nil
}
