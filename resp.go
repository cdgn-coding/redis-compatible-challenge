package main

import (
	"errors"
	"fmt"
)

var EmptyError = errors.New("cannot serialize empty error")

var SimpleStringsError = errors.New("cannot serialize simple strings with \\n or \\r")

var RespNil = "$-1\r\n"

type RespSerializer struct{}

func (RespSerializer) SerializeString(data string) ([]byte, error) {
	for c := range data {
		if c == '\n' || c == '\r' {
			return nil, SimpleStringsError
		}
	}

	result := fmt.Sprintf("+%s\r\n", data)
	return []byte(result), nil
}

func (RespSerializer) SerializeInteger(data int64) ([]byte, error) {
	result := fmt.Sprintf(":%d\r\n", data)
	return []byte(result), nil
}

func (RespSerializer) SerializeError(data error) ([]byte, error) {
	if data == nil {
		return []byte{}, EmptyError
	}
	result := fmt.Sprintf("-%s\r\n", data)
	return []byte(result), nil
}

func (RespSerializer) SerializeBulkString(data string) ([]byte, error) {
	result := fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
	return []byte(result), nil
}
