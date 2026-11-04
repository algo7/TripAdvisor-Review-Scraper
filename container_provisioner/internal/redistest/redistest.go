// Package redistest connects tests to a real Redis server.
package redistest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
)

// New returns a client for the Redis server at TEST_REDIS_ADDR. When it is not set, the test fails
// in CI and is skipped elsewhere.
func New(t *testing.T) *database.RedisClient {
	t.Helper()

	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_REDIS_ADDR is not set")
		}
		t.Skip("TEST_REDIS_ADDR is not set. Start Redis with `docker run --rm -p 6379:6379 redis:8.10.3-alpine` and set TEST_REDIS_ADDR=localhost:6379")
	}

	t.Setenv("REDIS_HOST", addr)
	t.Setenv("REDIS_PASS", "")
	r := database.NewRedisClient()
	t.Cleanup(func() { r.Client.Close() })

	_, err := r.CheckConnection()
	if err != nil {
		t.Fatalf("Redis at %s: %v", addr, err)
	}

	return r
}

// UniqueID returns a random hex string, so tests never share keys with each other or with
// existing data in the Redis server.
func UniqueID(t *testing.T) string {
	t.Helper()

	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	return hex.EncodeToString(b)
}

// Cleanup deletes the given keys when the test ends.
func Cleanup(t *testing.T, r *database.RedisClient, keys ...string) {
	t.Helper()
	// t.Context() is already cancelled when cleanup functions run
	t.Cleanup(func() { r.Client.Del(context.Background(), keys...) })
}
