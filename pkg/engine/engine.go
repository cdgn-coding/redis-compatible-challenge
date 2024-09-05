package engine

import (
	"errors"
	concurrency2 "github.com/cdgn-coding/redis-compatible-challenge/pkg/concurrency"
	"reflect"
	"strconv"
)

var UnsupportedCommandError = errors.New("unsupported command")

var UnsupportedTypeForCommand = errors.New("unsupported type")

func ConcurrentListConstructor() interface{} {
	return concurrency2.NewConcurrentList()
}

type Engine struct {
	memory *concurrency2.ConcurrentMap
}

func NewEngine() *Engine {
	return &Engine{
		memory: concurrency2.NewConcurrentMap(),
	}
}

const COMMAND = "COMMAND"
const PING = "PING"
const ECHO = "ECHO"
const GET = "GET"
const SET = "SET"
const DEL = "DEL"
const EXISTS = "EXISTS"
const INCR = "INCR"
const DECR = "DECR"
const RPUSH = "RPUSH"
const LPUSH = "LPUSH"

var DOCS = []interface{}{}

const OK = "OK"

func (e *Engine) Process(payload interface{}) (interface{}, error) {
	if reflect.TypeOf(payload).Kind() != reflect.Slice {
		return nil, UnsupportedCommandError
	}

	payloadArray := payload.([]interface{})
	firstPart := payloadArray[0].(string)
	switch firstPart {
	case COMMAND:
		command := payloadArray[1].(string)
		switch command {
		case "DOCS":
			return DOCS, nil
		default:
			return nil, UnsupportedCommandError
		}
	case PING:
		return []interface{}{"PONG"}, nil
	case ECHO:
		return payloadArray[1:], nil
	case GET:
		key := payloadArray[1].(string)
		val, ok := e.memory.Get(key)

		if !ok {
			return nil, nil
		}

		return val, nil
	case SET:
		key := payloadArray[1].(string)
		val := payloadArray[2]
		e.memory.Set(key, val)
		return OK, nil
	case DEL:
		for _, key := range payloadArray[1:] {
			e.memory.Delete(key.(string))
		}

		if len(payloadArray) == 2 {
			return OK, nil
		}

		return int64(len(payloadArray) - 1), nil
	case EXISTS:
		var count int64 = 0
		for _, key := range payloadArray[1:] {
			if e.memory.Has(key.(string)) {
				count++
			}
		}

		return count, nil
	case INCR:
		key := payloadArray[1].(string)
		err := e.memory.Map(key, e.incrementMapper)
		if err != nil {
			return nil, err
		}
		return OK, nil
	case DECR:
		key := payloadArray[1].(string)
		err := e.memory.Map(key, e.incrementMapper)
		if err != nil {
			return nil, err
		}
		return OK, nil
	case RPUSH:
		key := payloadArray[1].(string)
		var val interface{}
		var err error
		for _, newValue := range payloadArray[2:] {
			val, err = e.memory.Mutate(key, e.pushRight(newValue), ConcurrentListConstructor)
			if err != nil {
				return nil, err
			}
		}
		return val, nil
	case LPUSH:
		key := payloadArray[1].(string)
		var val interface{}
		var err error
		for _, newValue := range payloadArray[2:] {
			val, err = e.memory.Mutate(key, e.pushLeft(newValue), ConcurrentListConstructor)
			if err != nil {
				return nil, err
			}
		}
		return val, nil
	default:
		return nil, UnsupportedCommandError
	}
}

func (e *Engine) pushRight(newValue interface{}) concurrency2.MapperFunc {
	return func(val interface{}) (interface{}, error) {
		if concurrency2.IsConcurrentList(val) {
			return nil, UnsupportedTypeForCommand
		}

		list := val.(*concurrency2.ConcurrentList)
		list.PushRight(newValue)
		return list.Len(), nil
	}
}

func (e *Engine) pushLeft(newValue interface{}) concurrency2.MapperFunc {
	return func(val interface{}) (interface{}, error) {
		if concurrency2.IsConcurrentList(val) {
			return nil, UnsupportedTypeForCommand
		}

		list := val.(*concurrency2.ConcurrentList)
		list.PushLeft(newValue)
		return list.Len(), nil
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
