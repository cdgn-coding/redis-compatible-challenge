package resp

import (
	"errors"
	"fmt"
	"reflect"
)

var EmptyError = errors.New("cannot serialize empty error")

var SimpleStringsError = errors.New("cannot serialize simple strings with \\n or \\r")

var UnknownType = errors.New("cannot serialize unknown type")

var ArrayError = errors.New("cannot serialize array")

var RespNull = "$-1\r\n"

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

func (s RespSerializer) SerializeArray(data []interface{}) ([]byte, error) {
	result := []byte(fmt.Sprintf("*%d\r\n", len(data)))
	var err error
	var part []byte
	// var err error

	for _, element := range data {
		t := reflect.TypeOf(element)

		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			part, err = s.SerializeArray(element.([]interface{}))
		case reflect.Int:
			part, err = s.SerializeInteger(int64(element.(int)))
		case reflect.String:
			part, err = s.SerializeBulkString(element.(string))
		default:
			return nil, UnknownType
		}

		if err != nil {
			return nil, errors.Join(err, ArrayError)
		}

		result = append(result, part...)
	}

	return result, nil
}
