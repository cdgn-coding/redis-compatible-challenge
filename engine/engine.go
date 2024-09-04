package engine

import (
	"errors"
	"reflect"
	"sync"
)

var UnsupportedCommandError = errors.New("unsupported command")

type Entry struct {
	value interface{}
	lock  sync.Mutex
}

func NewEntry(value interface{}) Entry {
	return Entry{
		value: value,
		lock:  sync.Mutex{},
	}
}

func (e *Entry) Read() interface{} {
	return e.value
}

func (e *Entry) Write(value interface{}) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.value = value
}

type ConcurrentMap struct {
	memory map[string]Entry
}

func NewConcurrentMap() ConcurrentMap {
	return ConcurrentMap{
		memory: make(map[string]Entry),
	}
}

func (c *ConcurrentMap) Set(key string, value interface{}) {
	entry, ok := c.memory[key]
	if !ok {
		c.memory[key] = NewEntry(value)
		return
	}

	entry.Write(value)
}

func (c *ConcurrentMap) Get(key string) (interface{}, bool) {
	entry, ok := c.memory[key]
	if !ok {
		return nil, false
	}

	return entry.Read(), true
}

type Engine struct {
	memory ConcurrentMap
}

func NewEngine() *Engine {
	return &Engine{
		memory: NewConcurrentMap(),
	}
}

func (e *Engine) Process(payload interface{}) (interface{}, error) {
	if reflect.TypeOf(payload).Kind() != reflect.Slice {
		return nil, UnsupportedCommandError
	}

	payloadArray := payload.([]interface{})
	firstPart := payloadArray[0].(string)

	switch firstPart {
	case "COMMAND":
		command := payloadArray[1].(string)
		switch command {
		case "DOCS":
			return []interface{}{}, nil
		default:
			return nil, UnsupportedCommandError
		}
	case "PING":
		return []interface{}{"PONG"}, nil
	case "ECHO":
		return payloadArray[1:], nil
	case "GET":
		key := payloadArray[1].(string)
		val, ok := e.memory.Get(key)
		if !ok {
			return nil, nil
		}

		return val, nil
	case "SET":
		key := payloadArray[1].(string)
		val := payloadArray[2]
		e.memory.Set(key, val)
		return "OK", nil
	default:
		return nil, UnsupportedCommandError
	}
}
