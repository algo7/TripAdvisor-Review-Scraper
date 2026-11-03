package database_test

import (
	"testing"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/internal/redistest"
)

func TestAcquireLockIsExclusive(t *testing.T) {
	r := redistest.New(t)
	key := "test-lock:" + redistest.UniqueID(t)
	redistest.Cleanup(t, r, key)

	acquired, err := r.AcquireLock(key, "a", time.Minute)
	if err != nil || !acquired {
		t.Fatalf("first acquire: acquired=%v err=%v, want true", acquired, err)
	}

	acquired, err = r.AcquireLock(key, "b", time.Minute)
	if err != nil || acquired {
		t.Fatalf("second acquire: acquired=%v err=%v, want false", acquired, err)
	}

	if owner := r.Client.Get(t.Context(), key).Val(); owner != "a" {
		t.Fatalf("owner = %q, want %q", owner, "a")
	}
	if ttl := r.Client.PTTL(t.Context(), key).Val(); ttl <= 50*time.Second || ttl > time.Minute {
		t.Fatalf("ttl = %v, want close to 1m", ttl)
	}
}

func TestRenewLockOnlyByOwner(t *testing.T) {
	r := redistest.New(t)
	key := "test-lock:" + redistest.UniqueID(t)
	redistest.Cleanup(t, r, key)

	r.AcquireLock(key, "a", time.Second)
	time.Sleep(300 * time.Millisecond)

	renewed, err := r.RenewLock(key, "b", time.Minute)
	if err != nil || renewed {
		t.Fatalf("renew by non-owner: renewed=%v err=%v, want false", renewed, err)
	}
	if ttl := r.Client.PTTL(t.Context(), key).Val(); ttl > 800*time.Millisecond {
		t.Fatalf("ttl after renew by non-owner = %v, want it unchanged (under 800ms)", ttl)
	}

	renewed, err = r.RenewLock(key, "a", time.Minute)
	if err != nil || !renewed {
		t.Fatalf("renew by owner: renewed=%v err=%v, want true", renewed, err)
	}
	if ttl := r.Client.PTTL(t.Context(), key).Val(); ttl <= 50*time.Second {
		t.Fatalf("ttl after renew by owner = %v, want close to 1m", ttl)
	}
}

func TestRenewLockAfterExpiry(t *testing.T) {
	r := redistest.New(t)
	key := "test-lock:" + redistest.UniqueID(t)
	redistest.Cleanup(t, r, key)

	r.AcquireLock(key, "a", 100*time.Millisecond)
	time.Sleep(250 * time.Millisecond)

	renewed, err := r.RenewLock(key, "a", time.Minute)
	if err != nil || renewed {
		t.Fatalf("renew after expiry: renewed=%v err=%v, want false", renewed, err)
	}
	if r.Client.Exists(t.Context(), key).Val() != 0 {
		t.Fatal("renewing an expired lock must not recreate it")
	}
}

func TestReleaseLockOnlyByOwner(t *testing.T) {
	r := redistest.New(t)
	key := "test-lock:" + redistest.UniqueID(t)
	redistest.Cleanup(t, r, key)

	r.AcquireLock(key, "a", time.Minute)

	released, err := r.ReleaseLock(key, "b")
	if err != nil || released {
		t.Fatalf("release by non-owner: released=%v err=%v, want false", released, err)
	}
	if r.Client.Exists(t.Context(), key).Val() != 1 {
		t.Fatal("release by non-owner deleted the lock")
	}

	released, err = r.ReleaseLock(key, "a")
	if err != nil || !released {
		t.Fatalf("release by owner: released=%v err=%v, want true", released, err)
	}
	if r.Client.Exists(t.Context(), key).Val() != 0 {
		t.Fatal("release by owner left the lock in place")
	}
}
