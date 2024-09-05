package engine

import (
	"errors"
	"reflect"
	"strconv"
)

var UnsupportedCommandError = errors.New("unsupported command")

var UnsupportedTypeForCommand = errors.New("unsupported type")

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
	case "INCR":
		key := payloadArray[1].(string)
		err := e.memory.Map(key, e.incrementMapper)
		if err != nil {
			return nil, err
		}
		return "OK", nil
	case "DECR":
		key := payloadArray[1].(string)
		err := e.memory.Map(key, e.incrementMapper)
		if err != nil {
			return nil, err
		}
		return "OK", nil
	default:
		return nil, UnsupportedCommandError
	}
}

func (e *Engine) decrementMapper(val interface{}) (interface{}, error) {
	if val == nil {
		return int64(-1), nil
	}

	t := reflect.TypeOf(val)

	if t.Kind() == reflect.Int64 {
		return val.(int64) - 1, nil
	}

	if t.Kind() != reflect.String {
		return 0, UnsupportedTypeForCommand
	}

	numberVal, err := strconv.ParseInt(val.(string), 10, 64)
	if err != nil {
		return 0, err
	}

	return numberVal - 1, nil
}

func (e *Engine) incrementMapper(val interface{}) (interface{}, error) {
	if val == nil {
		return int64(0), nil
	}

	t := reflect.TypeOf(val)

	if t.Kind() == reflect.Int64 {
		return val.(int64) + 1, nil
	}

	if t.Kind() != reflect.String {
		return 0, UnsupportedTypeForCommand
	}

	numberVal, err := strconv.ParseInt(val.(string), 10, 64)
	if err != nil {
		return 0, err
	}

	return numberVal + 1, nil
}
