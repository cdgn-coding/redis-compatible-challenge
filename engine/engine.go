package engine

import (
	"errors"
	"reflect"
)

var UnsupportedCommandError = errors.New("unsupported command")

type Engine struct{}

func (e Engine) Process(payload interface{}) ([]interface{}, error) {
	if reflect.TypeOf(payload).Kind() != reflect.Slice {
		return nil, UnsupportedCommandError
	}

	payloadArray := payload.([]interface{})
	firstPart := string(payloadArray[0].([]byte))

	switch firstPart {
	case "COMMAND":
		command := string(payloadArray[1].([]byte))
		switch command {
		case "DOCS":
			return []interface{}{}, nil
		default:
			return nil, UnsupportedCommandError
		}
	case "PING":
		return []interface{}{"PONG"}, nil
	default:
		return nil, UnsupportedCommandError
	}
}
