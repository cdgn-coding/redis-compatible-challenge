package resp

import (
	"bytes"
	"errors"
	"testing"
)

func TestRespSerializer_Serialize(t *testing.T) {
	cases := []struct {
		data     interface{}
		expected []byte
		err      error
		message  string
	}{
		{
			data:     127,
			expected: []byte(":127\r\n"),
			message:  "Positive integer",
			err:      nil,
		},
		{
			data:     errors.New("unknown error"),
			expected: []byte("-unknown error\r\n"),
			message:  "Non empty error",
			err:      nil,
		},
		{
			data:     nil,
			expected: []byte(RespNull),
			message:  "Null data",
			err:      nil,
		},
		{
			data:     []interface{}{},
			expected: []byte("*0\r\n"),
			message:  "EmptyPayload array",
			err:      nil,
		},
		{
			data: []interface{}{
				"hello",
				"world",
			},
			expected: []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
			message:  "Array two bulk strings",
			err:      nil,
		},
		{
			data:     []interface{}{1, 2, 3},
			expected: []byte("*3\r\n:1\r\n:2\r\n:3\r\n"),
			message:  "Array of three integers",
			err:      nil,
		},
		{
			data:     []interface{}{1, 2, 3, 4, "hello"},
			expected: []byte("*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n"),
			message:  "Array of mixed data types",
			err:      nil,
		},
		{
			data:     []interface{}{[]interface{}{1, 2, 3, 4}, []interface{}{"hello", "world"}},
			expected: []byte("*2\r\n*4\r\n:1\r\n:2\r\n:3\r\n:4\r\n*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
			message:  "Array of arrays",
			err:      nil,
		},
		{
			data:     []interface{}{errors.New("some error")},
			expected: []byte("*1\r\n-some error\r\n"),
			message:  "Array with error",
			err:      nil,
		},
	}

	serializer := RespSerializer{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			actual, err := serializer.Serialize(c.data)
			if !errors.Is(err, c.err) || bytes.Compare(actual.Bytes(), c.expected) != 0 {
				t.Errorf("Actual data %s, actual error %v, expected data %s, expected error %v", actual, err, c.expected, c.err)
			}
		})
	}
}
