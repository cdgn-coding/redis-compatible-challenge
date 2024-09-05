package concurrency

import (
	"sync"
	"testing"
)

func TestConcurrentList_PushLeft(t *testing.T) {
	tests := []struct {
		name     string
		elements []interface{}
		sequence []interface{}
	}{
		{
			name:     "No elements",
			elements: []interface{}{},
			sequence: []interface{}{},
		},
		{
			name:     "One elements",
			elements: []interface{}{1},
			sequence: []interface{}{1},
		},
		{
			name:     "Multiple elements",
			elements: []interface{}{1, 2, 3, 4, 5},
			sequence: []interface{}{5, 4, 3, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewConcurrentList()
			for _, el := range tt.elements {
				cl.PushLeft(el)
			}

			i := 0
			for el := range cl.Iterator() {
				if tt.sequence[i] != el {
					t.Errorf("got %v, want %v", el, tt.sequence[i])
				}
				i++
			}
		})
	}
}

func TestConcurrentList_PushRight(t *testing.T) {
	tests := []struct {
		name     string
		elements []interface{}
		sequence []interface{}
	}{
		{
			name:     "No elements",
			elements: []interface{}{},
			sequence: []interface{}{},
		},
		{
			name:     "One elements",
			elements: []interface{}{1},
			sequence: []interface{}{1},
		},
		{
			name:     "Multiple elements",
			elements: []interface{}{1, 2, 3, 4, 5},
			sequence: []interface{}{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewConcurrentList()
			for _, el := range tt.elements {
				cl.PushRight(el)
			}

			i := 0
			for el := range cl.Iterator() {
				if tt.sequence[i] != el {
					t.Errorf("got %v, want %v", el, tt.sequence[i])
				}
				i++
			}
		})
	}
}

func TestConcurrentPushes(t *testing.T) {
	list := NewConcurrentList()
	var wg sync.WaitGroup
	const numGoroutines = 50
	const numPushesPerGoroutine = 100

	// Test PushLeft
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numPushesPerGoroutine; j++ {
				list.PushLeft(i*100 + j)
			}
		}(i)
	}
	wg.Wait()

	if list.Len() != numGoroutines*numPushesPerGoroutine {
		t.Errorf("Expected list length %d, got %d", numGoroutines*numPushesPerGoroutine, list.Len())
	}

	// Reset list for next test
	list = NewConcurrentList()

	// Test PushRight
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numPushesPerGoroutine; j++ {
				list.PushRight(i*100 + j)
			}
		}(i)
	}
	wg.Wait()

	if list.Len() != numGoroutines*numPushesPerGoroutine {
		t.Errorf("Expected list length %d, got %d", numGoroutines*numPushesPerGoroutine, list.Len())
	}
}

func TestConcurrentIterationAndModification(t *testing.T) {
	list := NewConcurrentList()
	// Pre-populate the list
	for i := 0; i < 100; i++ {
		list.PushRight(i)
	}

	var wg sync.WaitGroup
	const numGoroutines = 10

	// Start goroutines that modify the list
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			list.PushRight(i)
		}(i)
	}

	// Start goroutines that iterate over the list
	results := make(chan []interface{}, numGoroutines)
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			var result []interface{}
			for val := range list.Iterator() {
				result = append(result, val)
			}
			results <- result
		}()
	}

	wg.Wait()
	close(results)

	// Check that all contain the first elements
	for result := range results {
		if len(result) < 100 {
			t.Errorf("Iterator missed elements, got %d elements, want at least 100", len(result))
		}
	}
}
