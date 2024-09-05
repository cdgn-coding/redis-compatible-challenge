package engine

import (
	"errors"
	"fmt"
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

func TestConcurrentSet(t *testing.T) {
	cm := NewConcurrentMap()

	var wg sync.WaitGroup
	numGoroutines := 100
	key := "key"

	// Concurrently set the same key with different values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cm.Set(key, i)
		}(i)
	}

	wg.Wait()

	// Check if the final value of the key is set (it should be one of the goroutines' values)
	value, ok := cm.Get(key)
	if !ok {
		t.Errorf("expected key %s to exist, but it was not found", key)
	}

	// Since we are setting the same key concurrently, the final value can be from any goroutine.
	// We just need to make sure that it exists and has a valid value (within the range of goroutines).
	finalValue := value.(int)
	if finalValue < 0 || finalValue >= numGoroutines {
		t.Errorf("unexpected value for key %s: got %d, want value between 0 and %d", key, finalValue, numGoroutines-1)
	}
}

func TestConcurrentGet(t *testing.T) {
	cm := NewConcurrentMap()

	// Set up a key-value pair before concurrent access
	key := "key"
	expectedValue := "value"
	cm.Set(key, expectedValue)

	var wg sync.WaitGroup
	numGoroutines := 100
	errs := make(chan error, numGoroutines)

	// Concurrently get the same key from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, ok := cm.Get(key)
			if !ok || value != expectedValue {
				errs <- errors.New(fmt.Sprintf("expected value %s for key %s, but got %v", expectedValue, key, value))
			}
		}()
	}

	wg.Wait()
	close(errs)

	// If any errors occurred during concurrent reads, fail the test
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func incrementMapper(v interface{}) (interface{}, error) {
	if v == nil {
		return 1, nil
	}

	if val, ok := v.(int); ok {
		return val + 1, nil
	}
	return v, nil
}

// decrementMapper decrements the value by 1
func decrementMapper(v interface{}) (interface{}, error) {
	if v == nil {
		return -1, nil
	}

	if val, ok := v.(int); ok {
		return val - 1, nil
	}
	return v, nil
}

func TestConcurrentMapWithIncrementDecrement(t *testing.T) {
	cm := NewConcurrentMap()

	var wg sync.WaitGroup
	numGoroutines := 10

	cm.Set("counter", 0)

	// Increment and decrement the same key concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				cm.Map("counter", incrementMapper) // Increment
			} else {
				cm.Map("counter", decrementMapper) // Decrement
			}
		}(i)
	}

	wg.Wait()

	// Check the final value of the counter
	value, ok := cm.Get("counter")
	if !ok {
		t.Fatalf("expected key 'counter' to exist, but it was not found")
	}

	finalValue := value.(int)

	// Since we have an equal number of increments and decrements, the final value should be 0.
	if finalValue != 0 {
		t.Errorf("expected final value to be 0, but got %d", finalValue)
	}
}
