package engine

import (
	"testing"
)

func BenchmarkEngine_Process_LPUSH(b *testing.B) {
	eng := NewEngine()
	b.SetParallelism(2)

	b.RunParallel(func(pb *testing.PB) {
		command := []interface{}{
			"LPUSH",
			"value",
		}
		for pb.Next() {
			_, err := eng.Process(command)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
