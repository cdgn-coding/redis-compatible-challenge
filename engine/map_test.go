package engine

import (
	"sync"
	"testing"
)

func TestConcurrentDelete(t *testing.T) {
	cm := NewConcurrentMap()

	// Set up keys
	keys := []string{"key1", "key2", "key3", "key4"}
	for _, key := range keys {
		cm.Set(key, "value")
	}

	var wg sync.WaitGroup

	// Run multiple goroutines to delete keys concurrently
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			cm.Delete(k)
		}(key)
	}

	// Run multiple goroutines to read keys concurrently
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			_, _ = cm.Get(k)
		}(key)
	}

	wg.Wait()

	// After all deletes, ensure that all keys are gone
	for _, key := range keys {
		val, _ := cm.Get(key)
		if val != nil {
			t.Errorf("expected key %s to be deleted", key)
		}
	}
}
