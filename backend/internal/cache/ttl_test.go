package cache

import (
	"testing"
	"time"
)

func TestTTLCacheExpiresValues(t *testing.T) {
	cache := NewTTL[string, string](10 * time.Millisecond)
	cache.Set("key", "value")

	if value, ok := cache.Get("key"); !ok || value != "value" {
		t.Fatalf("Get before expiry = %q, %v; want value, true", value, ok)
	}

	time.Sleep(20 * time.Millisecond)
	if _, ok := cache.Get("key"); ok {
		t.Fatalf("Get after expiry returned cached value")
	}
}

func TestTTLCacheDelete(t *testing.T) {
	cache := NewTTL[string, int](time.Minute)
	cache.Set("key", 1)
	cache.Delete("key")

	if _, ok := cache.Get("key"); ok {
		t.Fatalf("Get after Delete returned cached value")
	}
}
