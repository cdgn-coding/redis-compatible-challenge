package engine

import (
	"errors"
	"fmt"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/concurrency"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/resp"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
)

var UnsupportedCommandError = errors.New("unsupported command")

var UnsupportedTypeForCommand = errors.New("unsupported type")

func ConcurrentListConstructor() interface{} {
	return concurrency.NewConcurrentList()
}

type Engine struct {
	memory     *concurrency.ConcurrentMap
	serializer *resp.RespSerializer
	parser     *resp.RespParser
	file       string
	global     bool
}

type EngineOptions struct {
	File       *string
	Load       *bool
	GlobalPath *bool
}

func NewEngine(opts EngineOptions) (*Engine, error) {
	eng := &Engine{
		memory:     concurrency.NewConcurrentMap(),
		serializer: &resp.RespSerializer{},
		parser:     &resp.RespParser{},
	}

	if opts.File != nil {
		eng.file = *opts.File
	} else {
		eng.file = filepath.Join(os.TempDir(), "memory.resp")
	}

	if opts.GlobalPath != nil && *opts.GlobalPath {
		eng.global = true
	}

	if opts.Load != nil && *opts.Load {
		err := eng.load()
		if err != nil {
			return nil, err
		}
	}

	return eng, nil
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
const SAVE = "SAVE"

var DOCS = []interface{}{}

const OK = "OK"
const PONG = "PONG"

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
		return PONG, nil
	case ECHO:
		if len(payloadArray) != 2 {
			return nil, UnsupportedTypeForCommand
		}
		return payloadArray[1], nil
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
		kind := reflect.TypeOf(val).Kind()
		if kind != reflect.Slice && kind != reflect.Array {
			e.memory.Set(key, val)
			return OK, nil
		}
		e.memory.Set(key, concurrency.NewConcurrentListFromSlice(val.([]interface{})))
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
		err := e.memory.Map(key, e.decrementMapper)
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
	case SAVE:
		err := e.save()
		if err != nil {
			return nil, err
		}
		return OK, nil
	default:
		return nil, UnsupportedCommandError
	}
}

func (e *Engine) load() error {
	savePath, err := e.getPath()
	if err != nil {
		return err
	}

	// Check if the File exists
	_, err = os.Stat(savePath)
	if os.IsNotExist(err) {
		// File doesn't exist, which is not an error
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check File status: %w", err)
	}

	// Open the File
	file, err := os.Open(filepath.Clean(savePath))
	if err != nil {
		return fmt.Errorf("failed to open File: %w", err)
	}
	defer file.Close()
	scanner := e.parser.CreateScanner(file)
	for result := range e.parser.Iterate(scanner) {
		if result.Err() != nil {
			return result.Err()
		}

		_, err = e.Process(result.Value())
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) getPath() (string, error) {
	if e.global {
		return e.file, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	savePath := filepath.Join(dir, e.file)

	return savePath, nil
}

func (e *Engine) save() error {
	savePath, err := e.getPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	saveDir := filepath.Dir(savePath)
	err = os.MkdirAll(saveDir, 0750)
	if err != nil {
		return err
	}

	// Remove the existing File if it exists
	err = os.Remove(savePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create the new File
	file, err := os.Create(filepath.Clean(savePath))
	if err != nil {
		return err
	}
	defer file.Close()

	for pair := range e.memory.Iterable() {
		command := []interface{}{SET, pair.Key, pair.Value}
		payload, err := e.serializer.Serialize(command)
		if err != nil {
			return err
		}

		_, err = file.Write(payload.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) pushRight(newValue interface{}) concurrency.MapperFunc {
	return func(val interface{}) (interface{}, error) {
		ls, ok := val.(*concurrency.ConcurrentList)
		if !ok {
			return nil, UnsupportedTypeForCommand
		}
		ls.PushRight(newValue)
		return ls.Len(), nil
	}
}

func (e *Engine) pushLeft(newValue interface{}) concurrency.MapperFunc {
	return func(val interface{}) (interface{}, error) {
		ls, ok := val.(*concurrency.ConcurrentList)
		if !ok {
			return nil, UnsupportedTypeForCommand
		}
		ls.PushLeft(newValue)
		return ls.Len(), nil
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
		return int64(1), nil
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
