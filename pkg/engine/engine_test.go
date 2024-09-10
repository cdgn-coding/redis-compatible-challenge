package engine

import (
	"testing"
)

func BenchmarkEngine_Process_LPUSH(b *testing.B) {
	eng, _ := NewEngine(EngineOptions{})

	command := []interface{}{
		"LPUSH",
		"value",
	}
	for n := 0; n < b.N; n++ {
		_, err := eng.Process(command)
		if err != nil {
			b.Fatal(err)
		}
	}
}
