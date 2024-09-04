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
	default:
		return nil, UnsupportedCommandError
	}
}
