package utils

import (
	"testing"
	"time"
)

func TestCacheRace(t *testing.T) {
	cache := NewCache[string, string](10 * time.Millisecond)

	done := make(chan bool)
	go func() {
		for i := 0; i < 1000; i++ {
			cache.Set("key", "value")
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			cache.Get("key")
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	<-done
	<-done
}

func BenchmarkCacheGetContention(b *testing.B) {
	cache := NewCache[string, string](time.Minute)
	cache.Set("key", "value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val, ok := cache.Get("key")
			if !ok || val != "value" {
				b.Fatal("invalid value")
			}
		}
	})
}
