package engine

import (
	"errors"
	"reflect"
)

var UnsupportedCommandError = errors.New("unsupported command")

type Engine struct {
	memory *ConcurrentMap
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
	case "DEL":
		for _, key := range payloadArray[1:] {
			e.memory.Delete(key.(string))
		}
		return "OK", nil
	default:
		return nil, UnsupportedCommandError
	}
}
