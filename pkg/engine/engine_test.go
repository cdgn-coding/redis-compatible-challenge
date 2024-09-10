package engine

import (
	"fmt"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/concurrency"
	"io"
	"os"
	"strings"
	"testing"
)

func TestEngine_Process(t *testing.T) {
	temp := t.TempDir()
	data := fmt.Sprintf("%s/data.resp", temp)
	seedDataFile := fmt.Sprintf("%s/exampleData.resp", temp)
	exampleData := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nhello\r\n*3\r\n$3\r\nSET\r\n$4\r\nkey2\r\n$5\r\nworld\r\n"

	file, err := os.Create(seedDataFile)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write([]byte(exampleData))
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	tt := []struct {
		dataFile *string
		load     bool
		name     string
		assert   func(engine *Engine) bool
	}{
		{
			name: "Set & GET key",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("SET key hello"))
				res, _ := eng.Process(toCommand("GET key"))
				return res.(string) == "hello"
			},
		},
		{
			name: "GET key that don't exist",
			assert: func(eng *Engine) bool {
				res, err := eng.Process(toCommand("GET key"))
				return res == nil && err == nil
			},
		},
		{
			name: "PING",
			assert: func(eng *Engine) bool {
				res, err := eng.Process(toCommand("PING"))
				return res.(string) == "PONG" && err == nil
			},
		},
		{
			name: "ECHO",
			assert: func(eng *Engine) bool {
				res, err := eng.Process(toCommand("ECHO hello-world!"))
				return res.(string) == "hello-world!" && err == nil
			},
		},
		{
			name: "DEL existing key",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("SET key hello"))
				eng.Process(toCommand("DEL key"))
				res, err := eng.Process(toCommand("GET key"))
				return res == nil && err == nil
			},
		},
		{
			name: "DEL non existing key",
			assert: func(eng *Engine) bool {
				res, err := eng.Process(toCommand("DEL key"))
				return err == nil && res.(string) == "OK"
			},
		},
		{
			name: "EXISTS non existing key",
			assert: func(eng *Engine) bool {
				res, err := eng.Process(toCommand("EXISTS key"))
				return err == nil && res.(int64) == 0
			},
		},
		{
			name: "EXISTS existing key",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("SET key hello"))
				res, err := eng.Process(toCommand("EXISTS key"))
				return err == nil && res.(int64) == 1
			},
		},
		{
			name: "EXISTS multiple keys",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("SET key1 hello"))
				eng.Process(toCommand("SET key2 hello"))
				eng.Process(toCommand("SET key3 hello"))
				res, err := eng.Process(toCommand("EXISTS key1 key2 key3 key4"))
				return err == nil && res.(int64) == 3
			},
		},
		{
			name: "INCR once",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("INCR counter"))
				res, err := eng.Process(toCommand("GET counter"))
				return err == nil && res.(int64) == 1
			},
		},
		{
			name: "INCR multiple times",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("INCR counter"))
				eng.Process(toCommand("INCR counter"))
				eng.Process(toCommand("INCR counter"))
				res, err := eng.Process(toCommand("GET counter"))
				return err == nil && res.(int64) == 3
			},
		},
		{
			name: "DECR once",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("DECR counter"))
				res, err := eng.Process(toCommand("GET counter"))
				return err == nil && res.(int64) == -1
			},
		},
		{
			name: "DECR multiple times",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("DECR counter"))
				eng.Process(toCommand("DECR counter"))
				eng.Process(toCommand("DECR counter"))
				res, err := eng.Process(toCommand("GET counter"))
				return err == nil && res.(int64) == -3
			},
		},
		{
			name: "RPUSH respond size of list",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("RPUSH arr 1"))
				eng.Process(toCommand("RPUSH arr 2"))
				res, err := eng.Process(toCommand("RPUSH arr 3"))
				return err == nil && res.(int) == 3
			},
		},
		{
			name: "RPUSH stores a list",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("RPUSH arr 1"))
				eng.Process(toCommand("RPUSH arr 2"))
				eng.Process(toCommand("RPUSH arr 3"))
				res, err := eng.Process(toCommand("GET arr"))
				if err != nil {
					return false
				}

				cl := res.(*concurrency.ConcurrentList)
				return cl.Len() == 3
			},
		},
		{
			name: "LPUSH respond size of list",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("LPUSH arr 1"))
				eng.Process(toCommand("LPUSH arr 2"))
				res, err := eng.Process(toCommand("LPUSH arr 3"))
				return err == nil && res.(int) == 3
			},
		},
		{
			name: "LPUSH stores a list",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("LPUSH arr 1"))
				eng.Process(toCommand("LPUSH arr 2"))
				eng.Process(toCommand("LPUSH arr 3"))
				res, err := eng.Process(toCommand("GET arr"))
				if err != nil {
					return false
				}

				cl := res.(*concurrency.ConcurrentList)
				return cl.Len() == 3
			},
		},
		{
			dataFile: &data,
			name:     "SAVE",
			assert: func(eng *Engine) bool {
				eng.Process(toCommand("SET key hello"))
				eng.Process(toCommand("SET key2 world"))
				resp, err := eng.Process(toCommand("SAVE"))
				if err != nil && resp.(string) != "OK" {
					return false
				}

				file, _ := os.Open(data)
				defer file.Close()
				bytes, _ := io.ReadAll(file)
				return string(bytes) == exampleData
			},
		},
		{
			load:     true,
			dataFile: &seedDataFile,
			name:     "LOAD",
			assert: func(eng *Engine) bool {
				resp, err := eng.Process(toCommand("GET key"))
				if err != nil || resp.(string) != "hello" {
					return false
				}
				resp, err = eng.Process(toCommand("GET key2"))
				return err == nil || resp.(string) == "world"
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			global := true
			eng, _ := NewEngine(EngineOptions{
				File:       tc.dataFile,
				Load:       &tc.load,
				GlobalPath: &global,
			})
			if !tc.assert(eng) {
				t.Fail()
			}
		})
	}
}

func toCommand(command string) []interface{} {
	data := strings.Split(command, " ")

	payload := make([]interface{}, len(data))

	for i := range data {
		payload[i] = data[i]
	}

	return payload
}

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
